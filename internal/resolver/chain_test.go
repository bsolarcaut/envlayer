package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/envlayer/internal/resolver"
)

// stubResolver is a simple in-memory resolver used for testing.
type stubResolver struct {
	name string
	data map[string]string
	err  error
}

func (s *stubResolver) Name() string { return s.name }

func (s *stubResolver) Resolve(_ context.Context, key string) (string, bool, error) {
	if s.err != nil {
		return "", false, s.err
	}
	v, ok := s.data[key]
	return v, ok, nil
}

func TestChain_Resolve_FirstWins(t *testing.T) {
	a := &stubResolver{name: "a", data: map[string]string{"FOO": "from-a"}}
	b := &stubResolver{name: "b", data: map[string]string{"FOO": "from-b", "BAR": "from-b"}}
	chain := resolver.NewChain(a, b)

	val, ok, err := chain.Resolve(context.Background(), "FOO")
	if err != nil || !ok || val != "from-a" {
		t.Fatalf("expected from-a, got %q ok=%v err=%v", val, ok, err)
	}

	val, ok, err = chain.Resolve(context.Background(), "BAR")
	if err != nil || !ok || val != "from-b" {
		t.Fatalf("expected from-b, got %q ok=%v err=%v", val, ok, err)
	}
}

func TestChain_Resolve_NotFound(t *testing.T) {
	chain := resolver.NewChain(&stubResolver{name: "a", data: map[string]string{}})
	_, ok, err := chain.Resolve(context.Background(), "MISSING")
	if err != nil || ok {
		t.Fatalf("expected not found, got ok=%v err=%v", ok, err)
	}
}

func TestChain_Resolve_PropagatesError(t *testing.T) {
	expected := errors.New("backend unavailable")
	chain := resolver.NewChain(&stubResolver{name: "broken", err: expected})
	_, _, err := chain.Resolve(context.Background(), "ANY")
	if err == nil || !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestChain_MustResolve_ErrorOnMissing(t *testing.T) {
	chain := resolver.NewChain(&stubResolver{name: "empty", data: map[string]string{}})
	_, err := chain.MustResolve(context.Background(), "SECRET")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestChain_Names(t *testing.T) {
	chain := resolver.NewChain(
		&stubResolver{name: "dotenv"},
		&stubResolver{name: "ssm"},
	)
	names := chain.Names()
	if len(names) != 2 || names[0] != "dotenv" || names[1] != "ssm" {
		t.Fatalf("unexpected names: %v", names)
	}
}
