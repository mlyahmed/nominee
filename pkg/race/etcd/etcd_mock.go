package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"time"
)

type MockServerStatistics struct {
	ConnectHits        int
	NewElectionHits    int
	ResumeElectionHits int
	StopHits           int
	CleanupHits        int
}

// MockServerConnector ...
type MockServerConnector struct {
	StopChan         chan struct{}
	Client           *MockClient
	Statistics       *MockServerStatistics
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
}

// NewMockServerConnector ...
func NewMockServerConnector() *MockServerConnector {
	mock := &MockServerConnector{
		Statistics: &MockServerStatistics{},
	}

	mock.ConnectFn = func(ctx context.Context, config *etcdconfig.Config) (Client, error) {
		return mock.Client, nil
	}

	mock.NewElectionFn = func(context.Context, string) (Election, error) {
		return mock.Election, nil
	}

	mock.ResumeElectionFn = func(context.Context, string, clientv3.GetResponse) (Election, error) {
		return mock.Election, nil
	}

	mock.StopFn = func() nominee.StopChan {
		return mock.StopChan
	}

	mock.CleanupFn = func() {
		select {
		case <-mock.StopChan:
			// Make sure the chan is closed
		default:
			close(mock.StopChan)
		}
	}

	return mock
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
			<-ctx.Done()
			return nil
		},
		ObserveFn: func(ctx context.Context) <-chan clientv3.GetResponse {
			return leaderChan
		},
	}
}

// Connect ...
func (mock *MockServerConnector) Connect(ctx context.Context, config *etcdconfig.Config) (Client, error) {
	mock.StopChan = make(chan struct{}, 1)
	mock.Client = NewMockClient()
	mock.Statistics.ConnectHits++
	return mock.ConnectFn(ctx, config)
}

// NewElection ...
func (mock *MockServerConnector) NewElection(ctx context.Context, electionKey string) (Election, error) {
	mock.Election = NewMockElection()
	mock.Statistics.NewElectionHits++
	return mock.NewElectionFn(ctx, electionKey)
}

// ResumeElection ...
func (mock *MockServerConnector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error) {
	mock.Election = NewMockElection()
	mock.Statistics.ResumeElectionHits++
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

// CloseSession ...
func (mock *MockServerConnector) CloseSession() {
	mock.StopChan <- struct{}{}
	time.Sleep(10 * time.Millisecond)
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
	go func() {
		<-ctx.Done()
		close(mock.LeaderChan)
	}()
	return mock.ObserveFn(ctx)
}

func (mock *MockElection) PushLeader(leader clientv3.GetResponse) {
	mock.LeaderChan <- leader
	time.Sleep(10 * time.Millisecond)
}
