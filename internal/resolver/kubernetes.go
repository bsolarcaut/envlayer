package resolver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const defaultSecretsMount = "/var/run/secrets"

// KubernetesResolver reads secrets from Kubernetes-mounted secret volumes.
// Each secret key is expected to be a file within the mount path.
type KubernetesResolver struct {
	mountPath string
	mu        sync.RWMutex
	cache     map[string]string
}

// NewKubernetesResolver returns a Resolver that reads from Kubernetes secret
// volume mounts. If mountPath is empty, it defaults to /var/run/secrets.
func NewKubernetesResolver(mountPath string) *KubernetesResolver {
	if mountPath == "" {
		mountPath = defaultSecretsMount
	}
	return &KubernetesResolver{
		mountPath: mountPath,
		cache:     make(map[string]string),
	}
}

// Name returns the resolver identifier.
func (r *KubernetesResolver) Name() string {
	return "kubernetes"
}

// Resolve looks up key as a file under the configured mount path.
// The key is lowercased and underscores are converted to hyphens to match
// typical Kubernetes secret naming conventions.
func (r *KubernetesResolver) Resolve(_ context.Context, key string) (string, error) {
	normalised := strings.ToLower(strings.ReplaceAll(key, "_", "-"))

	r.mu.RLock()
	if val, ok := r.cache[normalised]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	path := filepath.Join(r.mountPath, normalised)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("kubernetes: read %s: %w", path, err)
	}

	val := strings.TrimRight(string(data), "\n")

	r.mu.Lock()
	r.cache[normalised] = val
	r.mu.Unlock()

	return val, nil
}
