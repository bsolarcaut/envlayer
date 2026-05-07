package resolver

import (
	"context"
	"errors"
	"testing"
)

type mockConsulClient struct {
	data map[string]string
	err  error
	calls int
}

func (m *mockConsulClient) Get(_ context.Context, key string) (string, bool, error) {
	m.calls++
	if m.err != nil {
		return "", false, m.err
	}
	val, ok := m.data[key]
	return val, ok, nil
}

func TestConsulResolver_Resolve(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{"myapp/DB_PASSWORD": "s3cr3t"}}
	r := NewConsulResolver(client, "myapp/")

	val, err := r.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestConsulResolver_NotFound(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{}}
	r := NewConsulResolver(client, "myapp/")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestConsulResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("connection refused")
	client := &mockConsulClient{err: sentinel}
	r := NewConsulResolver(client, "")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestConsulResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{"API_KEY": "cached-value"}}
	r := NewConsulResolver(client, "")

	for i := 0; i < 3; i++ {
		val, err := r.Resolve(context.Background(), "API_KEY")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if val != "cached-value" {
			t.Errorf("call %d: expected %q, got %q", i, "cached-value", val)
		}
	}
	if client.calls != 1 {
		t.Errorf("expected 1 client call due to caching, got %d", client.calls)
	}
}

func TestConsulResolver_Name(t *testing.T) {
	r := NewConsulResolver(&mockConsulClient{}, "")
	if r.Name() != "consul" {
		t.Errorf("expected name %q, got %q", "consul", r.Name())
	}
}
