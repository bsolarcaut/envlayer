package resolver

import (
	"context"
	"fmt"
)

// ResolveAll resolves a slice of keys using the chain, returning a map of
// key -> value for every key that was found. Keys not found are omitted.
func (c *Chain) ResolveAll(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		val, ok, err := c.Resolve(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("resolving key %q: %w", key, err)
		}
		if ok {
			result[key] = val
		}
	}
	return result, nil
}

// MustResolve resolves a key and returns an error if the key is not found
// in any resolver in the chain.
func (c *Chain) MustResolve(ctx context.Context, key string) (string, error) {
	val, ok, err := c.Resolve(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("key %q not found in any resolver (%v)", key, c.Names())
	}
	return val, nil
}
