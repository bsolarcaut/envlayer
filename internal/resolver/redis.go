package resolver

import (
	"context"
	"errors"
	"sync"
)

// RedisClient is the interface for interacting with Redis.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
}

// RedisResolver resolves secrets from a Redis instance.
type RedisResolver struct {
	client RedisClient
	prefix string
	cache  map[string]string
	mu     sync.RWMutex
}

// NewRedisResolver creates a new RedisResolver with an optional key prefix.
func NewRedisResolver(client RedisClient, prefix string) *RedisResolver {
	return &RedisResolver{
		client: client,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

// Name returns the resolver name.
func (r *RedisResolver) Name() string {
	return "redis"
}

// Resolve looks up the given key in Redis, applying the prefix if set.
func (r *RedisResolver) Resolve(ctx context.Context, key string) (string, error) {
	lookup := key
	if r.prefix != "" {
		lookup = r.prefix + key
	}

	r.mu.RLock()
	if val, ok := r.cache[lookup]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	val, err := r.client.Get(ctx, lookup)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}

	r.mu.Lock()
	r.cache[lookup] = val
	r.mu.Unlock()

	return val, nil
}
