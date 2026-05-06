package resolver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// parameterStoreClient defines the interface for AWS SSM Parameter Store operations.
type parameterStoreClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// ParameterStoreResolver resolves secrets from AWS SSM Parameter Store with decryption support.
type ParameterStoreResolver struct {
	client    parameterStoreClient
	pathPrefix string
	cache     map[string]string
}

// NewParameterStoreResolver creates a new ParameterStoreResolver.
// pathPrefix is prepended to every key lookup (e.g. "/myapp/prod/").
func NewParameterStoreResolver(client parameterStoreClient, pathPrefix string) *ParameterStoreResolver {
	return &ParameterStoreResolver{
		client:    client,
		pathPrefix: pathPrefix,
		cache:     make(map[string]string),
	}
}

// Name returns the resolver identifier.
func (r *ParameterStoreResolver) Name() string {
	return "parameterstore"
}

// Resolve fetches a parameter by key from AWS SSM Parameter Store.
// It uses WithDecryption=true to support SecureString parameters.
func (r *ParameterStoreResolver) Resolve(ctx context.Context, key string) (string, error) {
	if val, ok := r.cache[key]; ok {
		return val, nil
	}

	paramName := r.pathPrefix + key

	out, err := r.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("%w: parameterstore: %s: %w", ErrNotFound, key, err)
	}

	if out.Parameter == nil || out.Parameter.Value == nil {
		return "", fmt.Errorf("%w: parameterstore: %s", ErrNotFound, key)
	}

	val := *out.Parameter.Value
	r.cache[key] = val
	return val, nil
}
