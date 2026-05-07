package resolver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeSecretFile(t *testing.T, dir, name, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0o600); err != nil {
		t.Fatalf("writeSecretFile: %v", err)
	}
}

func TestKubernetesResolver_Resolve(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "db-password", "s3cr3t\n")

	r := NewKubernetesResolver(dir)
	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("got %q, want %q", val, "s3cr3t")
	}
}

func TestKubernetesResolver_NotFound(t *testing.T) {
	r := NewKubernetesResolver(t.TempDir())
	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestKubernetesResolver_CachesOnSecondCall(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "api-token", "tok123")

	r := NewKubernetesResolver(dir)
	ctx := context.Background()

	first, err := r.Resolve(ctx, "API_TOKEN")
	if err != nil {
		t.Fatalf("first resolve: %v", err)
	}

	// Remove the file to prove second call uses cache.
	if err := os.Remove(filepath.Join(dir, "api-token")); err != nil {
		t.Fatalf("remove: %v", err)
	}

	second, err := r.Resolve(ctx, "API_TOKEN")
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if first != second {
		t.Errorf("cache miss: got %q, want %q", second, first)
	}
}

func TestKubernetesResolver_DefaultMountPath(t *testing.T) {
	r := NewKubernetesResolver("")
	if r.mountPath != defaultSecretsMount {
		t.Errorf("got %q, want %q", r.mountPath, defaultSecretsMount)
	}
}

func TestKubernetesResolver_Name(t *testing.T) {
	r := NewKubernetesResolver("")
	if r.Name() != "kubernetes" {
		t.Errorf("got %q, want %q", r.Name(), "kubernetes")
	}
}
