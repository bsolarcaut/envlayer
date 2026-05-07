package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/envlayer/envlayer/internal/resolver"
)

type mockOnePasswordClient struct {
	items map[string]string
	err   error
	calls int
}

func (m *mockOnePasswordClient) GetItem(_ context.Context, _, item string) (string, error) {
	m.calls++
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.items[item]
	if !ok {
		return "", errors.New("item not found")
	}
	return val, nil
}

func TestOnePasswordResolver_Resolve(t *testing.T) {
	client := &mockOnePasswordClient{
		items: map[string]string{"DB_PASSWORD": "s3cr3t"},
	}
	r := resolver.NewOnePasswordResolver(client, "my-vault")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected 's3cr3t', got %q", val)
	}
}

func TestOnePasswordResolver_NotFound(t *testing.T) {
	client := &mockOnePasswordClient{
		items: map[string]string{},
	}
	r := resolver.NewOnePasswordResolver(client, "my-vault")

	_, err := r.Resolve(context.Background(), "MISSING")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, resolver.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestOnePasswordResolver_PropagatesClientError(t *testing.T) {
	client := &mockOnePasswordClient{
		err: errors.New("connect timeout"),
	}
	r := resolver.NewOnePasswordResolver(client, "my-vault")

	_, err := r.Resolve(context.Background(), "any-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOnePasswordResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockOnePasswordClient{
		items: map[string]string{"API_KEY": "abc123"},
	}
	r := resolver.NewOnePasswordResolver(client, "my-vault")

	_, _ = r.Resolve(context.Background(), "API_KEY")
	_, _ = r.Resolve(context.Background(), "API_KEY")

	if client.calls != 1 {
		t.Errorf("expected 1 client call, got %d", client.calls)
	}
}

func TestOnePasswordResolver_Name(t *testing.T) {
	r := resolver.NewOnePasswordResolver(nil, "")
	if r.Name() != "1password" {
		t.Errorf("unexpected name: %s", r.Name())
	}
}
