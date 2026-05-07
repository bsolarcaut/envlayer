package resolver

import (
	"context"
	"fmt"
	"sync"
)

// ConsulClient defines the interface for interacting with HashiCorp Consul KV store.
type ConsulClient interface {
	Get(ctx context.Context, key string) (string, bool, error)
}

type consulResolver struct {
	client ConsulClient
	prefix string
	cache  map[string]string
	mu     sync.RWMutex
}

// NewConsulResolver creates a Resolver that fetches secrets from a Consul KV store.
// prefix is prepended to every key lookup (e.g. "myapp/secrets/").
func NewConsulResolver(client ConsulClient, prefix string) Resolver {
	return &consulResolver{
		client: client,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

func (r *consulResolver) Name() string {
	return "consul"
}

func (r *consulResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	fullKey := fmt.Sprintf("%s%s", r.prefix, key)
	val, found, err := r.client.Get(ctx, fullKey)
	if err != nil {
		return "", fmt.Errorf("consul: failed to get key %q: %w", fullKey, err)
	}
	if !found {
		return "", ErrNotFound
	}

	r.mu.Lock()
	r.cache[key] = val
	r.mu.Unlock()

	return val, nil
}
