package mock

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/pkg/node"
	"testing"
	"time"
)

//
type ConfigSpec struct {
	t *testing.T
	*etcd.ConfigSpec
}

// ServerRecord ...
type ServerRecord struct {
	ConnectHits        int
	NewElectionHits    int
	ResumeElectionHits int
	StopHits           int
	CleanupHits        int
}

// Connector ...
type Connector struct {
	t *testing.T
	*ServerRecord
	StopChan         chan struct{}
	Client           *Client
	Election         *Election
	ConnectFn        func(context.Context, *etcd.ConfigSpec) (etcd.Client, error)
	NewElectionFn    func(context.Context, string) (etcd.Election, error)
	ResumeElectionFn func(context.Context, string, clientv3.GetResponse) (etcd.Election, error)
	StopFn           func() node.StopChan
	CleanupFn        func()
}

// Client ...
type Client struct {
	WatchFn func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	GetFn   func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

type ElectionRecord struct {
	CampaignHits  int
	ObserveHits   int
	ElectionKey   string
	CampaignValue string
	Leader        clientv3.GetResponse
}

// Election ...
type Election struct {
	*ElectionRecord
	LeaderChan chan clientv3.GetResponse
	CampaignFn func(ctx context.Context, val string) error
	ObserveFn  func(ctx context.Context) <-chan clientv3.GetResponse
}

func NewConfigSpec(t *testing.T, spec *etcd.ConfigSpec) *ConfigSpec {
	return &ConfigSpec{
		t:          t,
		ConfigSpec: spec,
	}
}

// NewConnector ...
func NewConnector(_ *testing.T) *Connector {
	mock := &Connector{ServerRecord: &ServerRecord{}}

	mock.ConnectFn = func(ctx context.Context, config *etcd.ConfigSpec) (etcd.Client, error) {
		return mock.Client, nil
	}

	mock.NewElectionFn = func(context.Context, string) (etcd.Election, error) {
		return mock.Election, nil
	}

	mock.ResumeElectionFn = func(context.Context, string, clientv3.GetResponse) (etcd.Election, error) {
		return mock.Election, nil
	}

	mock.StopFn = func() node.StopChan {
		return mock.StopChan
	}

	mock.CleanupFn = func() {
	}

	return mock
}

// NewMockClient ...
func NewMockClient() *Client {
	return &Client{
		WatchFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
			return nil
		},
		GetFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return nil, nil
		},
	}
}

// NewMockElection ...
func NewMockElection() *Election {
	leaderChan := make(chan clientv3.GetResponse, 1)
	return &Election{
		LeaderChan:     leaderChan,
		ElectionRecord: &ElectionRecord{},
		CampaignFn: func(ctx context.Context, val string) error {
			<-ctx.Done()
			return nil
		},
		ObserveFn: func(ctx context.Context) <-chan clientv3.GetResponse {
			return leaderChan
		},
	}
}

// ConfigSpec mock the etcd.ConfigSpec.Load function
func (conf *ConfigSpec) Load(_ context.Context) {
	conf.Loaded = true
}

// Connect ...
func (mock *Connector) Connect(ctx context.Context, config *etcd.ConfigSpec) (etcd.Client, error) {
	mock.StopChan = make(chan struct{}, 1)
	mock.Client = NewMockClient()
	mock.ConnectHits++
	return mock.ConnectFn(ctx, config)
}

// NewElection ...
func (mock *Connector) NewElection(ctx context.Context, electionKey string) (etcd.Election, error) {
	mock.Election = NewMockElection()
	mock.Election.ElectionKey = electionKey
	mock.NewElectionHits++
	return mock.NewElectionFn(ctx, electionKey)
}

// ResumeElection ...
func (mock *Connector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (etcd.Election, error) {
	mock.Election = NewMockElection()
	mock.Election.ElectionKey = electionKey
	mock.Election.Leader = leader
	mock.ResumeElectionHits++
	return mock.ResumeElectionFn(ctx, electionKey, leader)
}

// StopChan ...
func (mock *Connector) Stop() node.StopChan {
	mock.StopHits++
	return mock.StopFn()
}

// Cleanup ...
func (mock *Connector) Cleanup() {
	mock.CleanupHits++
	mock.CleanupFn()
}

// CloseSession ...
func (mock *Connector) CloseSession() {
	close(mock.StopChan)
	time.Sleep(10 * time.Millisecond)
}

// Watch ...
func (mock *Client) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return mock.WatchFn(ctx, key, opts...)
}

// Get ...
func (mock *Client) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return mock.GetFn(ctx, key, opts...)
}

// Campaign ...
func (mock *Election) Campaign(ctx context.Context, val string) error {
	mock.CampaignHits++
	mock.CampaignValue = val
	return mock.CampaignFn(ctx, val)
}

// Observe ...
func (mock *Election) Observe(ctx context.Context) <-chan clientv3.GetResponse {
	mock.ObserveHits++
	go func() {
		<-ctx.Done()
		close(mock.LeaderChan)
	}()
	return mock.ObserveFn(ctx)
}

func (mock *Election) PushLeader(leader clientv3.GetResponse) {
	mock.LeaderChan <- leader
	time.Sleep(10 * time.Millisecond)
}
