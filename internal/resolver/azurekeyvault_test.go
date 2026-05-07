package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/envlayer/envlayer/internal/resolver"
)

type mockAzureKeyVaultClient struct {
	secrets map[string]string
	err     error
	calls   int
}

func (m *mockAzureKeyVaultClient) GetSecret(_ context.Context, name string) (string, error) {
	m.calls++
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.secrets[name]
	if !ok {
		return "", errors.New("secret not found")
	}
	return val, nil
}

func TestAzureKeyVaultResolver_Resolve(t *testing.T) {
	client := &mockAzureKeyVaultClient{
		secrets: map[string]string{"my-secret": "supersecret"},
	}
	r := resolver.NewAzureKeyVaultResolver(client, "https://myvault.vault.azure.net")

	val, err := r.Resolve(context.Background(), "my-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "supersecret" {
		t.Errorf("expected 'supersecret', got %q", val)
	}
}

func TestAzureKeyVaultResolver_NotFound(t *testing.T) {
	client := &mockAzureKeyVaultClient{
		secrets: map[string]string{},
	}
	r := resolver.NewAzureKeyVaultResolver(client, "https://myvault.vault.azure.net")

	_, err := r.Resolve(context.Background(), "missing-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, resolver.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAzureKeyVaultResolver_PropagatesClientError(t *testing.T) {
	client := &mockAzureKeyVaultClient{
		err: errors.New("network failure"),
	}
	r := resolver.NewAzureKeyVaultResolver(client, "https://myvault.vault.azure.net")

	_, err := r.Resolve(context.Background(), "any-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAzureKeyVaultResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockAzureKeyVaultClient{
		secrets: map[string]string{"cached-key": "cached-value"},
	}
	r := resolver.NewAzureKeyVaultResolver(client, "https://myvault.vault.azure.net")

	_, _ = r.Resolve(context.Background(), "cached-key")
	_, _ = r.Resolve(context.Background(), "cached-key")

	if client.calls != 1 {
		t.Errorf("expected 1 client call, got %d", client.calls)
	}
}

func TestAzureKeyVaultResolver_Name(t *testing.T) {
	r := resolver.NewAzureKeyVaultResolver(nil, "")
	if r.Name() != "azure-key-vault" {
		t.Errorf("unexpected name: %s", r.Name())
	}
}
