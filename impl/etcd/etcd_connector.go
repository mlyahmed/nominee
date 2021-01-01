package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// Client ...
type Client interface {
	// extracted from clientv3.Watcher
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan

	// extracted from clientv3.KV
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

// Election ...
type Election interface {
	Campaign(ctx context.Context, val string) error
	Observe(ctx context.Context) <-chan clientv3.GetResponse
}

// Connector ...
type Connector interface {
	Connect(ctx context.Context, config *Config) (Client, error)
	NewElection(ctx context.Context, electionKey string) (Election, error)
	ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error)
	Stop() <-chan struct{}
	Cleanup()
}

// DefaultConnector ...
type DefaultConnector struct {
	client  *clientv3.Client
	session *concurrency.Session
}

// NewDefaultConnector ...
func NewDefaultConnector() *DefaultConnector {
	return &DefaultConnector{}
}

// Connect ...
func (server *DefaultConnector) Connect(ctx context.Context, config *Config) (Client, error) {
	var err error
	logger.Infof("create new session. Endpoints %s", config.Endpoints)

	cfg := clientv3.Config{
		Context:   ctx,
		Endpoints: config.Endpoints,
	}
	if server.client, err = clientv3.New(cfg); err != nil {
		return nil, err
	}

	if server.session, err = concurrency.NewSession(server.client, concurrency.WithTTL(1), concurrency.WithContext(ctx)); err != nil {
		return nil, err
	}

	return server.client, nil
}

// NewElection ...
func (server *DefaultConnector) NewElection(_ context.Context, electionKey string) (Election, error) {
	election := concurrency.NewElection(server.session, electionKey)
	return election, nil
}

// ResumeElection ...
func (server *DefaultConnector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error) {
	election := concurrency.ResumeElection(server.session, electionKey, string(leader.Kvs[0].Key), leader.Kvs[0].CreateRevision)
	return election, nil
}

// StopChan ...
func (server *DefaultConnector) Stop() <-chan struct{} {
	return server.session.Done()
}

// Cleanup ...
func (server *DefaultConnector) Cleanup() {
	if server.client != nil {
		_ = server.client.Close()
	}

	if server.session != nil {
		_ = server.session.Close()
	}
}
