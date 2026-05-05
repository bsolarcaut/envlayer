package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient is the interface for AWS SSM Parameter Store operations.
type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// SSMResolver resolves environment variables from AWS SSM Parameter Store.
type SSMResolver struct {
	client SSMClient
	prefix string
	cache  map[string]string
	mu     sync.RWMutex
}

// NewSSMResolver creates a new SSMResolver with the given SSM client and key prefix.
// The prefix is prepended to each key when looking up parameters (e.g. "/myapp/prod/").
func NewSSMResolver(client SSMClient, prefix string) *SSMResolver {
	return &SSMResolver{
		client: client,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

// Name returns the resolver identifier.
func (s *SSMResolver) Name() string {
	return "ssm"
}

// Resolve looks up the given key in SSM Parameter Store, using the configured prefix.
// Results are cached in memory after the first successful fetch.
func (s *SSMResolver) Resolve(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	if val, ok := s.cache[key]; ok {
		s.mu.RUnlock()
		return val, nil
	}
	s.mu.RUnlock()

	paramName := s.prefix + key
	out, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("ssm: failed to resolve %q: %w", paramName, err)
	}

	if out.Parameter == nil || out.Parameter.Value == nil {
		return "", fmt.Errorf("ssm: parameter %q has no value", paramName)
	}

	val := *out.Parameter.Value

	s.mu.Lock()
	s.cache[key] = val
	s.mu.Unlock()

	return val, nil
}
