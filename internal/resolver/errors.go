package resolver

import "errors"

// ErrNotFound is returned by a Resolver when the requested key does not exist
// in that backend. The chain resolver uses this sentinel to continue to the
// next resolver in the chain rather than treating the absence as a hard
// failure.
var ErrNotFound = errors.New("secret not found")

// Resolver is the common interface implemented by every backend (dotenv, SSM,
// Vault, env, …). Resolve returns the plaintext secret value for key, or an
// error. When the key is simply absent the error must wrap ErrNotFound so
// callers can distinguish "not here" from "something broke".
type Resolver interface {
	// Name returns a short identifier used in logs and error messages.
	Name() string
	// Resolve fetches the value associated with key from this backend.
	Resolve(key string) (string, error)
}
