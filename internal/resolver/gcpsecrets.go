package resolver

import (
	"context"
	"fmt"
	"sync"
)

// GCPSecretManagerClient is the interface for GCP Secret Manager operations.
type GCPSecretManagerClient interface {
	AccessSecretVersion(ctx context.Context, name string) (string, error)
}

// GCPSecretsResolver resolves secrets from GCP Secret Manager.
type GCPSecretsResolver struct {
	client    GCPSecretManagerClient
	projectID string
	cache     map[string]string
	mu        sync.RWMutex
}

// NewGCPSecretsResolver creates a new GCPSecretsResolver for the given project.
func NewGCPSecretsResolver(client GCPSecretManagerClient, projectID string) *GCPSecretsResolver {
	return &GCPSecretsResolver{
		client:    client,
		projectID: projectID,
		cache:     make(map[string]string),
	}
}

// Name returns the resolver name.
func (r *GCPSecretsResolver) Name() string {
	return "gcp-secret-manager"
}

// Resolve looks up the secret value for the given key from GCP Secret Manager.
// The key is used as the secret name, and the latest version is fetched.
func (r *GCPSecretsResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	resourceName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", r.projectID, key)
	val, err := r.client.AccessSecretVersion(ctx, resourceName)
	if err != nil {
		return "", fmt.Errorf("%w: gcp secret %q: %w", ErrNotFound, key, err)
	}

	r.mu.Lock()
	r.cache[key] = val
	r.mu.Unlock()

	return val, nil
}
