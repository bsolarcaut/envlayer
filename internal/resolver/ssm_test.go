package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type mockSSMClient struct {
	params    map[string]string
	callCount map[string]int
	err       error
}

func (m *mockSSMClient) GetParameter(_ context.Context, input *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	name := aws.ToString(input.Name)
	if m.callCount == nil {
		m.callCount = make(map[string]int)
	}
	m.callCount[name]++
	val, ok := m.params[name]
	if !ok {
		return nil, errors.New("ParameterNotFound")
	}
	return &ssm.GetParameterOutput{
		Parameter: &types.Parameter{Value: aws.String(val)},
	}, nil
}

func TestSSMResolver_Resolve(t *testing.T) {
	client := &mockSSMClient{
		params: map[string]string{"/myapp/DB_PASSWORD": "supersecret"},
	}
	r := NewSSMResolver(client, "/myapp/")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "supersecret" {
		t.Errorf("expected %q, got %q", "supersecret", val)
	}
}

func TestSSMResolver_NotFound(t *testing.T) {
	client := &mockSSMClient{params: map[string]string{}}
	r := NewSSMResolver(client, "/myapp/")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestSSMResolver_PropagatesClientError(t *testing.T) {
	client := &mockSSMClient{err: errors.New("network failure")}
	r := NewSSMResolver(client, "/prod/")

	_, err := r.Resolve(context.Background(), "API_KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSSMResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockSSMClient{
		params: map[string]string{"/myapp/TOKEN": "abc123"},
	}
	r := NewSSMResolver(client, "/myapp/")

	for i := 0; i < 3; i++ {
		_, err := r.Resolve(context.Background(), "TOKEN")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	if client.callCount["/myapp/TOKEN"] != 1 {
		t.Errorf("expected 1 SSM call due to caching, got %d", client.callCount["/myapp/TOKEN"])
	}
}

func TestSSMResolver_Name(t *testing.T) {
	r := NewSSMResolver(nil, "")
	if r.Name() != "ssm" {
		t.Errorf("expected name %q, got %q", "ssm", r.Name())
	}
}
