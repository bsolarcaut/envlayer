package resolver

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client is the interface for the AWS S3 client methods used by S3Resolver.
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type s3Resolver struct {
	client S3Client
	bucket string
	prefix string
	cache  map[string]string
	mu     sync.RWMutex
}

// NewS3Resolver creates a Resolver that fetches secret values stored as
// objects in an S3 bucket. Each key is mapped to an object whose path is
// "<prefix><key>" inside the given bucket.
func NewS3Resolver(client S3Client, bucket, prefix string) Resolver {
	return &s3Resolver{
		client: client,
		bucket: bucket,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

func (r *s3Resolver) Name() string { return "s3" }

func (r *s3Resolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	if val, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return val, nil
	}
	r.mu.RUnlock()

	objectKey := r.prefix + key
	out, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("s3: get object %q: %w", objectKey, err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("s3: read body for %q: %w", objectKey, err)
	}

	val := strings.TrimSpace(string(data))

	r.mu.Lock()
	r.cache[key] = val
	r.mu.Unlock()

	return val, nil
}
