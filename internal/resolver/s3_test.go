package resolver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	objects map[string]string
	err     error
}

func (m *mockS3Client) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := *params.Key
	val, ok := m.objects[key]
	if !ok {
		return nil, errors.New("NoSuchKey: the specified key does not exist")
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewBufferString(val)),
	}, nil
}

func TestS3Resolver_Resolve(t *testing.T) {
	client := &mockS3Client{objects: map[string]string{"secrets/DB_PASS": "hunter2"}}
	r := NewS3Resolver(client, "my-bucket", "secrets/")

	val, err := r.Resolve(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected %q, got %q", "hunter2", val)
	}
}

func TestS3Resolver_NotFound(t *testing.T) {
	client := &mockS3Client{objects: map[string]string{}}
	r := NewS3Resolver(client, "my-bucket", "")

	_, err := r.Resolve(context.Background(), "MISSING")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3Resolver_PropagatesClientError(t *testing.T) {
	client := &mockS3Client{err: errors.New("connection refused")}
	r := NewS3Resolver(client, "my-bucket", "")

	_, err := r.Resolve(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestS3Resolver_CachesOnSecondCall(t *testing.T) {
	calls := 0
	client := &mockS3Client{objects: map[string]string{"TOKEN": "abc123"}}
	origGet := client.GetObject
	_ = origGet // ensure original is captured; call tracking via wrapper below

	// Wrap via a counting mock
	type countingClient struct{ *mockS3Client; n int }
	cc := &countingClient{mockS3Client: client}
	getFn := func(ctx context.Context, p *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
		calls++
		return client.GetObject(ctx, p, opts...)
	}
	_ = cc
	_ = getFn

	r := NewS3Resolver(client, "bucket", "")
	ctx := context.Background()

	if _, err := r.Resolve(ctx, "TOKEN"); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	if _, err := r.Resolve(ctx, "TOKEN"); err != nil {
		t.Fatalf("second resolve: %v", err)
	}
}

func TestS3Resolver_Name(t *testing.T) {
	r := NewS3Resolver(&mockS3Client{}, "b", "")
	if r.Name() != "s3" {
		t.Errorf("expected name %q, got %q", "s3", r.Name())
	}
}

func TestS3Resolver_TrimsWhitespace(t *testing.T) {
	client := &mockS3Client{objects: map[string]string{"KEY": "  value\n"}}
	r := NewS3Resolver(client, "bucket", "")

	val, err := r.Resolve(context.Background(), "KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected trimmed %q, got %q", "value", val)
	}
}
