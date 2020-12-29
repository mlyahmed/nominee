package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
)

type MockServerConnector struct {
	Client           *MockClient
	Election         *MockElection
	ConnectFn        func(context.Context, *etcdconfig.Config) (Client, error)
	NewElectionFn    func(context.Context, string) (Election, error)
	ResumeElectionFn func(context.Context, string, clientv3.GetResponse) (Election, error)
	StopChanFn       func() nominee.StopChan
	CleanupFn        func()
}

type MockClient struct {
	WatchFn func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	GetFn   func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

type MockElection struct {
	CampaignFn func(ctx context.Context, val string) error
	ObserveFn  func(ctx context.Context) <-chan clientv3.GetResponse
	ResignFn   func(ctx context.Context) (err error)
}

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
		StopChanFn: func() nominee.StopChan {
			return make(nominee.StopChan)
		},
		CleanupFn: func() {
		},
	}
}

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

func NewMockElection() *MockElection {
	return &MockElection{
		CampaignFn: func(ctx context.Context, val string) error {
			return nil
		},
		ObserveFn: func(ctx context.Context) <-chan clientv3.GetResponse {
			return make(chan clientv3.GetResponse)
		},
		ResignFn: func(ctx context.Context) (err error) {
			return nil
		},
	}
}

func (mock *MockServerConnector) Connect(ctx context.Context, config *etcdconfig.Config) (Client, error) {
	return mock.ConnectFn(ctx, config)
}

func (mock *MockServerConnector) NewElection(ctx context.Context, electionKey string) (Election, error) {
	return mock.NewElectionFn(ctx, electionKey)
}

func (mock *MockServerConnector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error) {
	return mock.ResumeElectionFn(ctx, electionKey, leader)
}

func (mock *MockServerConnector) StopChan() nominee.StopChan {
	return mock.StopChanFn()
}

func (mock *MockServerConnector) Cleanup() {
	mock.CleanupFn()
}

func (mock *MockClient) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return mock.WatchFn(ctx, key, opts...)
}

func (mock *MockClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return mock.GetFn(ctx, key, opts...)
}

func (mock *MockElection) Campaign(ctx context.Context, val string) error {
	return mock.CampaignFn(ctx, val)
}

func (mock *MockElection) Observe(ctx context.Context) <-chan clientv3.GetResponse {
	return mock.ObserveFn(ctx)
}

func (mock *MockElection) Resign(ctx context.Context) (err error) {
	return mock.ResignFn(ctx)
}
