package resolver

import (
	"errors"
	"os"
	"testing"
)

func TestEnvResolver_Resolve(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Cleanup(func() { os.Unsetenv("DATABASE_URL") })

	r := NewEnvResolver("")
	val, err := r.Resolve("DATABASE_URL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "postgres://localhost/test" {
		t.Errorf("got %q, want %q", val, "postgres://localhost/test")
	}
}

func TestEnvResolver_ResolveWithPrefix(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", "supersecret")
	t.Cleanup(func() { os.Unsetenv("APP_SECRET_KEY") })

	r := NewEnvResolver("APP_")
	val, err := r.Resolve("SECRET_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "supersecret" {
		t.Errorf("got %q, want %q", val, "supersecret")
	}
}

func TestEnvResolver_NotFound(t *testing.T) {
	r := NewEnvResolver("")
	_, err := r.Resolve("DEFINITELY_NOT_SET_XYZ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEnvResolver_NotFoundWithPrefix(t *testing.T) {
	r := NewEnvResolver("MYAPP_")
	_, err := r.Resolve("MISSING_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEnvResolver_Name(t *testing.T) {
	r := NewEnvResolver("")
	if r.Name() != "env" {
		t.Errorf("got name %q, want %q", r.Name(), "env")
	}
}

func TestEnvResolver_EmptyValueIsValid(t *testing.T) {
	os.Setenv("EMPTY_VAR", "")
	t.Cleanup(func() { os.Unsetenv("EMPTY_VAR") })

	r := NewEnvResolver("")
	val, err := r.Resolve("EMPTY_VAR")
	if err != nil {
		t.Fatalf("unexpected error for empty env var: %v", err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}
