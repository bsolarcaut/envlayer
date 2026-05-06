package resolver

import (
	"context"
	"errors"
	"testing"
)

type mockGCPClient struct {
	secrets map[string]string
	err     error
	calls   int
}

func (m *mockGCPClient) AccessSecretVersion(_ context.Context, name string) (string, error) {
	m.calls++
	if m.err != nil {
		return "", m.err
	}
	if val, ok := m.secrets[name]; ok {
		return val, nil
	}
	return "", errors.New("secret not found")
}

func TestGCPSecretsResolver_Resolve(t *testing.T) {
	client := &mockGCPClient{
		secrets: map[string]string{
			"projects/my-project/secrets/DB_PASSWORD/versions/latest": "supersecret",
		},
	}
	r := NewGCPSecretsResolver(client, "my-project")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "supersecret" {
		t.Errorf("expected %q, got %q", "supersecret", val)
	}
}

func TestGCPSecretsResolver_NotFound(t *testing.T) {
	client := &mockGCPClient{secrets: map[string]string{}}
	r := NewGCPSecretsResolver(client, "my-project")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGCPSecretsResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("gcp unavailable")
	client := &mockGCPClient{err: sentinel}
	r := NewGCPSecretsResolver(client, "my-project")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error in chain, got %v", err)
	}
}

func TestGCPSecretsResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockGCPClient{
		secrets: map[string]string{
			"projects/proj/secrets/API_KEY/versions/latest": "cached-value",
		},
	}
	r := NewGCPSecretsResolver(client, "proj")

	if _, err := r.Resolve(context.Background(), "API_KEY"); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if _, err := r.Resolve(context.Background(), "API_KEY"); err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if client.calls != 1 {
		t.Errorf("expected 1 client call, got %d", client.calls)
	}
}

func TestGCPSecretsResolver_Name(t *testing.T) {
	r := NewGCPSecretsResolver(nil, "proj")
	if r.Name() != "gcp-secret-manager" {
		t.Errorf("unexpected name: %s", r.Name())
	}
}
