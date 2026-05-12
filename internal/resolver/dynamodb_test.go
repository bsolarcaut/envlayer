package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type mockDynamoDBClient struct {
	getItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	callCount int
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	m.callCount++
	return m.getItemFn(ctx, params, optFns...)
}

func TestDynamoDBResolver_Resolve(t *testing.T) {
	mock := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"value": &types.AttributeValueMemberS{Value: "supersecret"},
				},
			}, nil
		},
	}

	r := NewDynamoDBResolver(mock, "secrets", "key", "value")
	val, err := r.Resolve(context.Background(), "MY_SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "supersecret" {
		t.Errorf("expected %q, got %q", "supersecret", val)
	}
}

func TestDynamoDBResolver_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{Item: nil}, nil
		},
	}

	r := NewDynamoDBResolver(mock, "secrets", "key", "value")
	_, err := r.Resolve(context.Background(), "MISSING")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDynamoDBResolver_PropagatesClientError(t *testing.T) {
	sentinel := errors.New("connection refused")
	mock := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return nil, sentinel
		},
	}

	r := NewDynamoDBResolver(mock, "secrets", "key", "value")
	_, err := r.Resolve(context.Background(), "KEY")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestDynamoDBResolver_CachesOnSecondCall(t *testing.T) {
	mock := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"value": &types.AttributeValueMemberS{Value: "cached"},
				},
			}, nil
		},
	}

	r := NewDynamoDBResolver(mock, "secrets", "key", "value")
	ctx := context.Background()

	r.Resolve(ctx, "MY_KEY") //nolint:errcheck
	r.Resolve(ctx, "MY_KEY") //nolint:errcheck

	if mock.callCount != 1 {
		t.Errorf("expected 1 client call due to cache, got %d", mock.callCount)
	}
}

func TestDynamoDBResolver_Name(t *testing.T) {
	r := NewDynamoDBResolver(nil, "t", "k", "v")
	if r.Name() != "dynamodb" {
		t.Errorf("expected name %q, got %q", "dynamodb", r.Name())
	}
}
