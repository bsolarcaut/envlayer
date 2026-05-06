package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type mockSecretsManagerClient struct {
	output *secretsmanager.GetSecretValueOutput
	err    error
	calls  int
}

func (m *mockSecretsManagerClient) GetSecretValue(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	m.calls++
	return m.output, m.err
}

func TestSecretsManagerResolver_Resolve(t *testing.T) {
	mock := &mockSecretsManagerClient{
		output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String("s3cr3t")},
	}
	r := NewSecretsManagerResolver(mock, "myapp/")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestSecretsManagerResolver_NotFound(t *testing.T) {
	mock := &mockSecretsManagerClient{
		err: errors.New("ResourceNotFoundException"),
	}
	r := NewSecretsManagerResolver(mock, "")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSecretsManagerResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("network failure")
	mock := &mockSecretsManagerClient{err: sentinel}
	r := NewSecretsManagerResolver(mock, "")

	_, err := r.Resolve(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error in chain, got %v", err)
	}
}

func TestSecretsManagerResolver_CachesOnSecondCall(t *testing.T) {
	mock := &mockSecretsManagerClient{
		output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String("cached")},
	}
	r := NewSecretsManagerResolver(mock, "")

	for i := 0; i < 3; i++ {
		_, err := r.Resolve(context.Background(), "SOME_KEY")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 client call, got %d", mock.calls)
	}
}

func TestSecretsManagerResolver_Name(t *testing.T) {
	r := NewSecretsManagerResolver(nil, "")
	if r.Name() != "secretsmanager" {
		t.Errorf("expected name %q, got %q", "secretsmanager", r.Name())
	}
}
