//go:build integration
// +build integration

package resolver

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// TestDynamoDBResolver_Integration requires a real (or local) DynamoDB instance.
// Set DYNAMODB_TABLE, DYNAMODB_KEY_ATTR, DYNAMODB_VALUE_ATTR, and DYNAMODB_TEST_KEY
// environment variables before running.
func TestDynamoDBResolver_Integration(t *testing.T) {
	table := os.Getenv("DYNAMODB_TABLE")
	if table == "" {
		t.Skip("DYNAMODB_TABLE not set; skipping integration test")
	}

	keyAttr := os.Getenv("DYNAMODB_KEY_ATTR")
	if keyAttr == "" {
		keyAttr = "key"
	}

	valueAttr := os.Getenv("DYNAMODB_VALUE_ATTR")
	if valueAttr == "" {
		valueAttr = "value"
	}

	testKey := os.Getenv("DYNAMODB_TEST_KEY")
	if testKey == "" {
		t.Skip("DYNAMODB_TEST_KEY not set; skipping integration test")
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	r := NewDynamoDBResolver(client, table, keyAttr, valueAttr)

	val, err := r.Resolve(ctx, testKey)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", testKey, err)
	}
	if val == "" {
		t.Errorf("expected non-empty value for key %q", testKey)
	}
	t.Logf("resolved %q => %q", testKey, val)
}
