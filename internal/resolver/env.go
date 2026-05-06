package resolver

import (
	"fmt"
	"os"
)

// envResolver resolves secrets from the current process environment.
type envResolver struct {
	prefix string
}

// NewEnvResolver creates a Resolver that looks up keys from the process
// environment. An optional prefix is stripped from the key before lookup,
// so NewEnvResolver("APP_") will resolve "DATABASE_URL" by looking for
// the environment variable "APP_DATABASE_URL".
func NewEnvResolver(prefix string) Resolver {
	return &envResolver{prefix: prefix}
}

// Name returns a human-readable identifier for this resolver.
func (e *envResolver) Name() string {
	return "env"
}

// Resolve looks up key (with optional prefix prepended) in os.Environ.
// It returns ErrNotFound when the variable is absent so the chain can
// continue to the next resolver.
func (e *envResolver) Resolve(key string) (string, error) {
	lookup := e.prefix + key
	val, ok := os.LookupEnv(lookup)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrNotFound, lookup)
	}
	return val, nil
}
