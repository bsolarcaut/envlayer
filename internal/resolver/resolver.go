package resolver

import "context"

// Resolver is the interface that all secret/env backends must implement.
type Resolver interface {
	// Name returns a human-readable identifier for the resolver (e.g. "ssm", "vault", "dotenv").
	Name() string

	// Resolve attempts to retrieve the value for the given key.
	// Returns the value and nil on success.
	// Returns an empty string and a non-nil error if the key cannot be resolved.
	Resolve(ctx context.Context, key string) (string, error)
}

// ErrNotFound is returned by a Resolver when a key does not exist in the backend.
type ErrNotFound struct {
	Key      string
	Backend  string
}

func (e *ErrNotFound) Error() string {
	return e.Backend + ": key not found: " + e.Key
}

// chain holds an ordered list of Resolvers and tries each in sequence.
type chain struct {
	resolvers []Resolver
}

// NewChain returns a Resolver that tries each provided resolver in order,
// returning the first successful result. If no resolver finds the key,
// ErrNotFound is returned for the last resolver in the chain.
func NewChain(resolvers ...Resolver) Resolver {
	return &chain{resolvers: resolvers}
}

func (c *chain) Name() string {
	return "chain"
}

func (c *chain) Resolve(ctx context.Context, key string) (string, error) {
	for _, r := range c.resolvers {
		val, err := r.Resolve(ctx, key)
		if err == nil {
			return val, nil
		}
	}
	return "", &ErrNotFound{Key: key, Backend: "chain"}
}
