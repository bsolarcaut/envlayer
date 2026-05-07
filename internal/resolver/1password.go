package resolver

import (
	"context"
	"fmt"
	"sync"
)

// OnePasswordClient is the interface for fetching secrets from 1Password.
type OnePasswordClient interface {
	GetItem(ctx context.Context, vault, item string) (string, error)
}

// OnePasswordResolver resolves secrets from 1Password Connect.
type OnePasswordResolver struct {
	client OnePasswordClient
	vault  string
	cache  map[string]string
	mu     sync.RWMutex
}

// NewOnePasswordResolver creates a new OnePasswordResolver.
func NewOnePasswordResolver(client OnePasswordClient, vault string) *OnePasswordResolver {
	return &OnePasswordResolver{
		client: client,
		vault:  vault,
		cache:  make(map[string]string),
	}
}

// Name returns the resolver name.
func (r *OnePasswordResolver) Name() string {
	return "1password"
}

// Resolve fetches a secret item from 1Password by key.
func (r *OnePasswordResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	val, err := r.client.GetItem(ctx, r.vault, key)
	if err != nil {
		return "", fmt.Errorf("%w: 1password: %s", ErrNotFound, key)
	}

	r.mu.Lock()
	r.cache[key] = val
	r.mu.Unlock()

	return val, nil
}
