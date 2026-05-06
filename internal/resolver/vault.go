package resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// VaultClient is the interface for interacting with HashiCorp Vault.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// VaultResolver resolves environment variable values from HashiCorp Vault KV secrets.
type VaultResolver struct {
	client    VaultClient
	mountPath string
	mu        sync.Mutex
	cache     map[string]string
}

// NewVaultResolver creates a new VaultResolver using the provided client and KV mount path.
func NewVaultResolver(client VaultClient, mountPath string) *VaultResolver {
	return &VaultResolver{
		client:    client,
		mountPath: strings.TrimRight(mountPath, "/"),
		cache:     make(map[string]string),
	}
}

// Name returns the resolver identifier.
func (v *VaultResolver) Name() string {
	return "vault"
}

// Resolve looks up the given key in Vault. The key is used as the secret path
// relative to the mount path, and the value is read from the "value" field of
// the secret data. Returns ErrNotFound if the key is absent.
func (v *VaultResolver) Resolve(ctx context.Context, key string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if val, ok := v.cache[key]; ok {
		return val, nil
	}

	path := fmt.Sprintf("%s/%s", v.mountPath, key)
	data, err := v.client.ReadSecret(ctx, path)
	if err != nil {
		return "", fmt.Errorf("vault: reading secret at %q: %w", path, err)
	}

	if data == nil {
		return "", ErrNotFound
	}

	raw, ok := data["value"]
	if !ok {
		return "", ErrNotFound
	}

	val, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("vault: secret field \"value\" at %q is not a string", path)
	}

	v.cache[key] = val
	return val, nil
}
