package resolver

import (
	"context"
	"fmt"
	"sync"
)

// AzureKeyVaultClient is the interface for fetching secrets from Azure Key Vault.
type AzureKeyVaultClient interface {
	GetSecret(ctx context.Context, name string) (string, error)
}

// AzureKeyVaultResolver resolves secrets from Azure Key Vault.
type AzureKeyVaultResolver struct {
	client AzureKeyVaultClient
	vaultURL string
	cache map[string]string
	mu sync.RWMutex
}

// NewAzureKeyVaultResolver creates a new AzureKeyVaultResolver.
func NewAzureKeyVaultResolver(client AzureKeyVaultClient, vaultURL string) *AzureKeyVaultResolver {
	return &AzureKeyVaultResolver{
		client:   client,
		vaultURL: vaultURL,
		cache:    make(map[string]string),
	}
}

// Name returns the resolver name.
func (r *AzureKeyVaultResolver) Name() string {
	return "azure-key-vault"
}

// Resolve looks up the secret by key from Azure Key Vault.
func (r *AzureKeyVaultResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	val, err := r.client.GetSecret(ctx, key)
	if err != nil {
		return "", fmt.Errorf("%w: azure-key-vault: %s", ErrNotFound, key)
	}

	r.mu.Lock()
	r.cache[key] = val
	r.mu.Unlock()

	return val, nil
}
