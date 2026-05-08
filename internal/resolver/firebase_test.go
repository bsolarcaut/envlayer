package resolver

import (
	"context"
	"errors"
	"testing"
)

type mockFirebaseClient struct {
	values map[string]string
	err    error
	calls  int
}

func (m *mockFirebaseClient) GetRemoteConfigValue(_ context.Context, key string) (string, error) {
	m.calls++
	if m.err != nil {
		return "", m.err
	}
	return m.values[key], nil
}

func TestFirebaseResolver_Resolve(t *testing.T) {
	client := &mockFirebaseClient{values: map[string]string{"DB_HOST": "firebase-db.example.com"}}
	r := NewFirebaseResolver(client, "")

	val, err := r.Resolve(context.Background(), "DB_HOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "firebase-db.example.com" {
		t.Errorf("expected %q, got %q", "firebase-db.example.com", val)
	}
}

func TestFirebaseResolver_ResolveWithNamespace(t *testing.T) {
	client := &mockFirebaseClient{values: map[string]string{"prod/API_KEY": "secret-key"}}
	r := NewFirebaseResolver(client, "prod/")

	val, err := r.Resolve(context.Background(), "API_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret-key" {
		t.Errorf("expected %q, got %q", "secret-key", val)
	}
}

func TestFirebaseResolver_NotFound(t *testing.T) {
	client := &mockFirebaseClient{values: map[string]string{}}
	r := NewFirebaseResolver(client, "")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFirebaseResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("firebase unavailable")
	client := &mockFirebaseClient{err: sentinel}
	r := NewFirebaseResolver(client, "")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFirebaseResolver_CachesOnSecondCall(t *testing.T) {
	client := &mockFirebaseClient{values: map[string]string{"TOKEN": "abc123"}}
	r := NewFirebaseResolver(client, "")

	for i := 0; i < 3; i++ {
		_, err := r.Resolve(context.Background(), "TOKEN")
		if err != nil {
			t.Fatalf("call %d unexpected error: %v", i, err)
		}
	}
	if client.calls != 1 {
		t.Errorf("expected 1 client call due to caching, got %d", client.calls)
	}
}

func TestFirebaseResolver_Name(t *testing.T) {
	r := NewFirebaseResolver(&mockFirebaseClient{}, "")
	if r.Name() != "firebase" {
		t.Errorf("expected name %q, got %q", "firebase", r.Name())
	}
}
