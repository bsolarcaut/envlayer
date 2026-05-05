package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp env file: %v", err)
	}
	return path
}

func TestDotenvResolver_Resolve(t *testing.T) {
	path := writeTempEnvFile(t, `
# This is a comment
DB_HOST=localhost
DB_PORT=5432
API_KEY="super-secret"
SINGLE_QUOTED='hello world'
EMPTY_VAL=
`)

	r := NewDotenvResolver(path)

	tests := []struct {
		key      string
		wantVal  string
		wantErr  bool
	}{
		{"DB_HOST", "localhost", false},
		{"DB_PORT", "5432", false},
		{"API_KEY", "super-secret", false},
		{"SINGLE_QUOTED", "hello world", false},
		{"EMPTY_VAL", "", false},
		{"MISSING_KEY", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := r.Resolve(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
			if got != tt.wantVal {
				t.Errorf("Resolve(%q) = %q, want %q", tt.key, got, tt.wantVal)
			}
		})
	}
}

func TestDotenvResolver_FileNotFound(t *testing.T) {
	r := NewDotenvResolver("/nonexistent/path/.env")
	_, err := r.Resolve("ANY_KEY")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestDotenvResolver_Name(t *testing.T) {
	r := NewDotenvResolver(".env")
	if r.Name() != "dotenv" {
		t.Errorf("Name() = %q, want %q", r.Name(), "dotenv")
	}
}

func TestDotenvResolver_CachesOnSecondCall(t *testing.T) {
	path := writeTempEnvFile(t, "FOO=bar\n")
	r := NewDotenvResolver(path)

	// First call loads the file
	val1, err := r.Resolve("FOO")
	if err != nil || val1 != "bar" {
		t.Fatalf("first Resolve failed: val=%q err=%v", val1, err)
	}

	// Overwrite the file — cache should still return old value
	if err := os.WriteFile(path, []byte("FOO=changed\n"), 0600); err != nil {
		t.Fatal(err)
	}

	val2, err := r.Resolve("FOO")
	if err != nil || val2 != "bar" {
		t.Errorf("expected cached value %q, got %q (err=%v)", "bar", val2, err)
	}
}
