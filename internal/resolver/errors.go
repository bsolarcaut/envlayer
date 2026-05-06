package resolver

import "errors"

// ErrNotFound is returned by a Resolver when the requested key does not exist
// in the backing store. Callers should use errors.Is to check for this sentinel.
var ErrNotFound = errors.New("key not found")

// ErrResolverChainExhausted is returned by the Chain resolver when no resolver
// in the chain was able to provide a value for the requested key.
var ErrResolverChainExhausted = errors.New("resolver chain exhausted: key not found in any backend")
