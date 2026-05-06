package resolver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManagerClient defines the interface for AWS Secrets Manager operations.
type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// secretsManagerResolver resolves secrets from AWS Secrets Manager.
type secretsManagerResolver struct {
	client SecretsManagerClient
	prefix string
	cache  map[string]string
}

// NewSecretsManagerResolver creates a new resolver backed by AWS Secrets Manager.
// prefix is prepended to each key when constructing the secret ID.
func NewSecretsManagerResolver(client SecretsManagerClient, prefix string) Resolver {
	return &secretsManagerResolver{
		client: client,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

func (r *secretsManagerResolver) Name() string {
	return "secretsmanager"
}

func (r *secretsManagerResolver) Resolve(ctx context.Context, key string) (string, error) {
	if val, ok := r.cache[key]; ok {
		return val, nil
	}

	secretID := key
	if r.prefix != "" {
		secretID = fmt.Sprintf("%s%s", r.prefix, key)
	}

	out, err := r.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("%w: secretsmanager: %s: %w", ErrNotFound, key, err)
	}

	if out.SecretString == nil {
		return "", fmt.Errorf("%w: secretsmanager: %s: binary secrets are not supported", ErrNotFound, key)
	}

	val := aws.ToString(out.SecretString)
	r.cache[key] = val
	return val, nil
}
