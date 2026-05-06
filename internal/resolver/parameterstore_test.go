package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type mockParameterStoreClient struct {
	params map[string]string
	err    error
	calls  int
}

func (m *mockParameterStoreClient) GetParameter(_ context.Context, input *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	name := aws.ToString(input.Name)
	val, ok := m.params[name]
	if !ok {
		return nil, errors.New("ParameterNotFound")
	}
	return &ssm.GetParameterOutput{
		Parameter: &types.Parameter{
			Name:  aws.String(name),
			Value: aws.String(val),
		},
	}, nil
}

func TestParameterStoreResolver_Resolve(t *testing.T) {
	client := &mockParameterStoreClient{
		params: map[string]string{"/app/prod/DB_PASSWORD": "s3cr3t"},
	}
	r := NewParameterStoreResolver(client, "/app/prod/")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %s", val)
	}
}

func TestParameterStoreResolver_NotFound(t *testing.T) {
	client := &mockParameterStoreClient{
		params: map[string]string{},
	}
	r := NewParameterStoreResolver(client, "/app/prod/")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestParameterStoreResolver_PropagatesClientError(t *testing.T) {
	client := &mockParameterStoreClient{
		err: errors.New("network timeout"),
	}
	r := NewParameterStoreResolver(client, "/app/")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParameterStoreResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockParameterStoreClient{
		params: map[string]string{"/svc/API_KEY": "abc123"},
	}
	r := NewParameterStoreResolver(client, "/svc/")

	_, _ = r.Resolve(context.Background(), "API_KEY")
	_, _ = r.Resolve(context.Background(), "API_KEY")

	if client.calls != 1 {
		t.Errorf("expected 1 client call due to caching, got %d", client.calls)
	}
}

func TestParameterStoreResolver_Name(t *testing.T) {
	r := NewParameterStoreResolver(nil, "")
	if r.Name() != "parameterstore" {
		t.Errorf("expected parameterstore, got %s", r.Name())
	}
}
