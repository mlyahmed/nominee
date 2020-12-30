package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
)

// MockServerConnector ...
type MockServerConnector struct {
	Client           *MockClient
	Election         *MockElection
	ConnectFn        func(context.Context, *etcdconfig.Config) (Client, error)
	NewElectionFn    func(context.Context, string) (Election, error)
	ResumeElectionFn func(context.Context, string, clientv3.GetResponse) (Election, error)
	StopFn           func() nominee.StopChan
	CleanupFn        func()
}

// MockClient ...
type MockClient struct {
	WatchFn func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	GetFn   func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

// MockElection ...
type MockElection struct {
	LeaderChan chan clientv3.GetResponse
	CampaignFn func(ctx context.Context, val string) error
	ObserveFn  func(ctx context.Context) <-chan clientv3.GetResponse
	ResignFn   func(ctx context.Context) (err error)
}

// NewMockServerConnector ...
func NewMockServerConnector() *MockServerConnector {
	clientMock := NewMockClient()
	electionMock := NewMockElection()
	return &MockServerConnector{
		Client:   clientMock,
		Election: electionMock,
		ConnectFn: func(ctx context.Context, config *etcdconfig.Config) (Client, error) {
			return clientMock, nil
		},
		NewElectionFn: func(context.Context, string) (Election, error) {
			return electionMock, nil
		},
		ResumeElectionFn: func(context.Context, string, clientv3.GetResponse) (Election, error) {
			return electionMock, nil
		},
		StopFn: func() nominee.StopChan {
			return make(nominee.StopChan)
		},
		CleanupFn: func() {
		},
	}
}

// NewMockClient ...
func NewMockClient() *MockClient {
	return &MockClient{
		WatchFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
			return nil
		},
		GetFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return nil, nil
		},
	}
}

// NewMockElection ...
func NewMockElection() *MockElection {
	leaderChan := make(chan clientv3.GetResponse, 1)
	return &MockElection{
		LeaderChan: leaderChan,
		CampaignFn: func(ctx context.Context, val string) error {
			return nil
		},
		ObserveFn: func(ctx context.Context) <-chan clientv3.GetResponse {
			return leaderChan
		},
		ResignFn: func(ctx context.Context) (err error) {
			return nil
		},
	}
}

// Connect ...
func (mock *MockServerConnector) Connect(ctx context.Context, config *etcdconfig.Config) (Client, error) {
	return mock.ConnectFn(ctx, config)
}

// NewElection ...
func (mock *MockServerConnector) NewElection(ctx context.Context, electionKey string) (Election, error) {
	return mock.NewElectionFn(ctx, electionKey)
}

// ResumeElection ...
func (mock *MockServerConnector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error) {
	return mock.ResumeElectionFn(ctx, electionKey, leader)
}

// StopChan ...
func (mock *MockServerConnector) Stop() nominee.StopChan {
	return mock.StopFn()
}

// Cleanup ...
func (mock *MockServerConnector) Cleanup() {
	mock.CleanupFn()
}

// Watch ...
func (mock *MockClient) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return mock.WatchFn(ctx, key, opts...)
}

// Get ...
func (mock *MockClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return mock.GetFn(ctx, key, opts...)
}

// Campaign ...
func (mock *MockElection) Campaign(ctx context.Context, val string) error {
	return mock.CampaignFn(ctx, val)
}

// Observe ...
func (mock *MockElection) Observe(ctx context.Context) <-chan clientv3.GetResponse {
	return mock.ObserveFn(ctx)
}

// Resign ...
func (mock *MockElection) Resign(ctx context.Context) (err error) {
	return mock.ResignFn(ctx)
}
