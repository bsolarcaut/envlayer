package resolver

import (
	"context"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// etcdClient defines the interface used by EtcdResolver for fetching keys.
type etcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

// EtcdResolver resolves secrets from an etcd cluster.
type EtcdResolver struct {
	client etcdClient
	prefix string
	cache  map[string]string
	mu     sync.Mutex
}

// NewEtcdResolver creates a new EtcdResolver.
// prefix is an optional key prefix (e.g. "/secrets/").
func NewEtcdResolver(client etcdClient, prefix string) *EtcdResolver {
	return &EtcdResolver{
		client: client,
		prefix: prefix,
		cache:  make(map[string]string),
	}
}

// Name returns the resolver identifier.
func (r *EtcdResolver) Name() string {
	return "etcd"
}

// Resolve looks up key in etcd, prepending the configured prefix.
// Results are cached after the first successful lookup.
func (r *EtcdResolver) Resolve(ctx context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if val, ok := r.cache[key]; ok {
		return val, nil
	}

	fullKey := r.prefix + key
	resp, err := r.client.Get(ctx, fullKey)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", ErrNotFound
	}

	val := string(resp.Kvs[0].Value)
	r.cache[key] = val
	return val, nil
}
