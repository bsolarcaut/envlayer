package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient defines the interface for DynamoDB operations used by the resolver.
type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

type dynamoDBResolver struct {
	client    DynamoDBClient
	table     string
	keyAttr   string
	valueAttr string
	cache     map[string]string
	mu        sync.RWMutex
}

// NewDynamoDBResolver creates a Resolver that fetches secrets from an AWS DynamoDB table.
// keyAttr is the partition key attribute name, valueAttr is the attribute holding the secret value.
func NewDynamoDBResolver(client DynamoDBClient, table, keyAttr, valueAttr string) Resolver {
	return &dynamoDBResolver{
		client:    client,
		table:     table,
		keyAttr:   keyAttr,
		valueAttr: valueAttr,
		cache:     make(map[string]string),
	}
}

func (r *dynamoDBResolver) Name() string {
	return "dynamodb"
}

func (r *dynamoDBResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			r.keyAttr: &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return "", fmt.Errorf("dynamodb: GetItem failed for key %q: %w", key, err)
	}

	if out.Item == nil {
		return "", ErrNotFound
	}

	attr, ok := out.Item[r.valueAttr]
	if !ok {
		return "", ErrNotFound
	}

	sv, ok := attr.(*types.AttributeValueMemberS)
	if !ok {
		return "", fmt.Errorf("dynamodb: attribute %q is not a string type", r.valueAttr)
	}

	r.mu.Lock()
	r.cache[key] = sv.Value
	r.mu.Unlock()

	return sv.Value, nil
}
