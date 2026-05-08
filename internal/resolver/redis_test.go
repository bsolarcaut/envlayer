package resolver

import (
	"context"
	"errors"
	"testing"
)

type mockRedisClient struct {
	data     map[string]string
	callCount int
}

func (m *mockRedisClient) Get(_ context.Context, key string) (string, error) {
	m.callCount++
	val, ok := m.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return val, nil
}

func TestRedisResolver_Resolve(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{"MY_SECRET": "hunter2"}}
	r := NewRedisResolver(client, "")

	val, err := r.Resolve(context.Background(), "MY_SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected hunter2, got %s", val)
	}
}

func TestRedisResolver_ResolveWithPrefix(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{"app/MY_SECRET": "s3cr3t"}}
	r := NewRedisResolver(client, "app/")

	val, err := r.Resolve(context.Background(), "MY_SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %s", val)
	}
}

func TestRedisResolver_NotFound(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{}}
	r := NewRedisResolver(client, "")

	_, err := r.Resolve(context.Background(), "MISSING")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("connection refused")
	client := &mockRedisClient{}
	client.data = nil // force nil map lookup to panic — use a custom error instead

	errClient := &errorRedisClient{err: sentinel}
	r := NewRedisResolver(errClient, "")

	_, err := r.Resolve(context.Background(), "KEY")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

type errorRedisClient struct{ err error }

func (e *errorRedisClient) Get(_ context.Context, _ string) (string, error) {
	return "", e.err
}

func TestRedisResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{"TOKEN": "abc123"}}
	r := NewRedisResolver(client, "")

	_, _ = r.Resolve(context.Background(), "TOKEN")
	_, _ = r.Resolve(context.Background(), "TOKEN")

	if client.callCount != 1 {
		t.Errorf("expected 1 client call due to caching, got %d", client.callCount)
	}
}

func TestRedisResolver_Name(t *testing.T) {
	r := NewRedisResolver(&mockRedisClient{}, "")
	if r.Name() != "redis" {
		t.Errorf("expected name 'redis', got %s", r.Name())
	}
}
