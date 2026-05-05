package resolver

import "context"

// Resolver defines the interface for all secret/env backends.
type Resolver interface {
	// Name returns the identifier for this resolver (e.g. "dotenv", "ssm", "vault").
	Name() string

	// Resolve attempts to retrieve the value for the given key.
	// Returns the value and true if found, or empty string and false if not.
	Resolve(ctx context.Context, key string) (string, bool, error)
}

// Chain holds an ordered list of resolvers and tries each in sequence.
type Chain struct {
	resolvers []Resolver
}

// NewChain creates a new Chain with the provided resolvers in priority order.
// The first resolver to return a value wins.
func NewChain(resolvers ...Resolver) *Chain {
	return &Chain{resolvers: resolvers}
}

// Resolve walks the chain and returns the first successful match.
// If no resolver finds the key, it returns ("", false, nil).
func (c *Chain) Resolve(ctx context.Context, key string) (string, bool, error) {
	for _, r := range c.resolvers {
		val, ok, err := r.Resolve(ctx, key)
		if err != nil {
			return "", false, fmt.Errorf("resolver %q: %w", r.Name(), err)
		}
		if ok {
			return val, true, nil
		}
	}
	return "", false, nil
}

// Names returns the names of all registered resolvers in order.
func (c *Chain) Names() []string {
	names := make([]string, len(c.resolvers))
	for i, r := range c.resolvers {
		names[i] = r.Name()
	}
	return names
}
