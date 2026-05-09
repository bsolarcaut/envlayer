//go:build integration
// +build integration

package resolver_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yourorg/envlayer/internal/resolver"
)

// TestS3Resolver_Integration reads a real S3 object.
// Requires env vars: ENVLAYER_S3_BUCKET, ENVLAYER_S3_KEY, ENVLAYER_S3_EXPECTED.
func TestS3Resolver_Integration(t *testing.T) {
	bucket := os.Getenv("ENVLAYER_S3_BUCKET")
	key := os.Getenv("ENVLAYER_S3_KEY")
	expected := os.Getenv("ENVLAYER_S3_EXPECTED")

	if bucket == "" || key == "" {
		t.Skip("ENVLAYER_S3_BUCKET and ENVLAYER_S3_KEY must be set")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)
	r := resolver.NewS3Resolver(client, bucket, "")

	val, err := r.Resolve(context.Background(), key)
	if err != nil {
		t.Fatalf("resolve %q: %v", key, err)
	}

	if expected != "" && val != expected {
		t.Errorf("expected %q, got %q", expected, val)
	}
	t.Logf("resolved %q => %q", key, val)
}
