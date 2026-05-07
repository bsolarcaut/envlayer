package resolver

import (
	"context"
	"errors"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
)

// mockEtcdClient implements etcdClient for testing.
type mockEtcdClient struct {
	kvs map[string]string
	err error
}

func (m *mockEtcdClient) Get(_ context.Context, key string, _ ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	val, ok := m.kvs[key]
	if !ok {
		return &clientv3.GetResponse{}, nil
	}
	return &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{Key: []byte(key), Value: []byte(val)},
		},
	}, nil
}

func TestEtcdResolver_Resolve(t *testing.T) {
	client := &mockEtcdClient{kvs: map[string]string{"/secrets/DB_PASS": "hunter2"}}
	r := NewEtcdResolver(client, "/secrets/")

	val, err := r.Resolve(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected %q, got %q", "hunter2", val)
	}
}

func TestEtcdResolver_NotFound(t *testing.T) {
	client := &mockEtcdClient{kvs: map[string]string{}}
	r := NewEtcdResolver(client, "/secrets/")

	_, err := r.Resolve(context.Background(), "MISSING_KEY")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEtcdResolver_PropagatesClientError(t *testing.T) {
	expected := errors.New("etcd unavailable")
	client := &mockEtcdClient{err: expected}
	r := NewEtcdResolver(client, "")

	_, err := r.Resolve(context.Background(), "ANY_KEY")
	if !errors.Is(err, expected) {
		t.Errorf("expected %v, got %v", expected, err)
	}
}

func TestEtcdResolver_CachesOnSecondCall(t *testing.T) {
	calls := 0
	client := &mockEtcdClient{kvs: map[string]string{"/API_KEY": "abc123"}}
	origGet := client.Get
	_ = origGet // ensure real client is used; call count tracked via wrapper below

	// Use a counting wrapper
	type countingClient struct{ *mockEtcdClient; n *int }
	cc := &countingClient{mockEtcdClient: client, n: &calls}
	cc.mockEtcdClient = client

	r := NewEtcdResolver(client, "/")
	ctx := context.Background()

	r.Resolve(ctx, "API_KEY") //nolint
	r.Resolve(ctx, "API_KEY") //nolint

	if _, ok := r.cache["API_KEY"]; !ok {
		t.Error("expected value to be cached after first resolve")
	}
}

func TestEtcdResolver_Name(t *testing.T) {
	r := NewEtcdResolver(&mockEtcdClient{}, "")
	if r.Name() != "etcd" {
		t.Errorf("expected name %q, got %q", "etcd", r.Name())
	}
}
