package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/envlayer/envlayer/internal/resolver"
)

// mockVaultClient implements resolver.VaultClient for testing.
type mockVaultClient struct {
	data map[string]map[string]interface{}
	err  error
	calls int
}

func (m *mockVaultClient) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.data[path], nil
}

func TestVaultResolver_Resolve(t *testing.T) {
	client := &mockVaultClient{
		data: map[string]map[string]interface{}{
			"secret/DB_PASSWORD": {"value": "s3cr3t"},
		},
	}
	r := resolver.NewVaultResolver(client, "secret")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestVaultResolver_NotFound(t *testing.T) {
	client := &mockVaultClient{
		data: map[string]map[string]interface{}{},
	}
	r := resolver.NewVaultResolver(client, "secret")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if !errors.Is(err, resolver.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestVaultResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("vault unavailable")
	client := &mockVaultClient{err: sentinel}
	r := resolver.NewVaultResolver(client, "secret")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestVaultResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockVaultClient{
		data: map[string]map[string]interface{}{
			"kv/TOKEN": {"value": "abc123"},
		},
	}
	r := resolver.NewVaultResolver(client, "kv")

	for i := 0; i < 3; i++ {
		_, err := r.Resolve(context.Background(), "TOKEN")
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, err)
		}
	}
	if client.calls != 1 {
		t.Errorf("expected 1 client call due to caching, got %d", client.calls)
	}
}

func TestVaultResolver_Name(t *testing.T) {
	r := resolver.NewVaultResolver(&mockVaultClient{}, "secret")
	if r.Name() != "vault" {
		t.Errorf("expected name %q, got %q", "vault", r.Name())
	}
}
