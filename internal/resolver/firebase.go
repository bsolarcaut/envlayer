package resolver

import (
	"context"
	"fmt"
	"sync"
)

// FirebaseClient is the interface for fetching secrets from Firebase Remote Config or Realtime Database.
type FirebaseClient interface {
	GetRemoteConfigValue(ctx context.Context, key string) (string, error)
}

type firebaseResolver struct {
	client    FirebaseClient
	namespace string
	cache     map[string]string
	mu        sync.RWMutex
}

// NewFirebaseResolver returns a Resolver backed by Firebase.
// namespace is an optional prefix applied to every key lookup (e.g. "prod/").
func NewFirebaseResolver(client FirebaseClient, namespace string) Resolver {
	return &firebaseResolver{
		client:    client,
		namespace: namespace,
		cache:     make(map[string]string),
	}
}

func (r *firebaseResolver) Name() string {
	return "firebase"
}

func (r *firebaseResolver) Resolve(ctx context.Context, key string) (string, error) {
	lookupKey := key
	if r.namespace != "" {
		lookupKey = r.namespace + key
	}

	r.mu.RLock()
	if val, ok := r.cache[lookupKey]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	val, err := r.client.GetRemoteConfigValue(ctx, lookupKey)
	if err != nil {
		return "", fmt.Errorf("firebase: resolve %q: %w", lookupKey, err)
	}

	if val == "" {
		return "", ErrNotFound
	}

	r.mu.Lock()
	r.cache[lookupKey] = val
	r.mu.Unlock()

	return val, nil
}
