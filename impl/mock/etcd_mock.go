package mock

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/pkg/base"
	"testing"
	"time"
)

//
type ConfigSpec struct {
	t *testing.T
	*etcd.ConfigSpec
}

// ConnectorRecord ...
type ConnectorRecord struct {
	ConnectHits        int
	NewElectionHits    int
	ResumeElectionHits int
	StopHits           int
	CleanupHits        int
}

// Connector ...
type Connector struct {
	t *testing.T
	*ConnectorRecord
	stopChan         chan struct{}
	Client           *Client
	Election         *Election
	ConnectFn        func(context.Context, *etcd.ConfigSpec) (etcd.Client, error)
	NewElectionFn    func(context.Context, string) (etcd.Election, error)
	ResumeElectionFn func(context.Context, string, clientv3.GetResponse) (etcd.Election, error)
	StopFn           func() base.DoneChan
	CleanupFn        func()
}

// Client ...
type Client struct {
	WatchCan clientv3.WatchChan
	WatchFn  func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	GetFn    func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
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
	leaderChan chan clientv3.GetResponse
	CampaignFn func(ctx context.Context, val string) error
	ObserveFn  func(ctx context.Context) <-chan clientv3.GetResponse
}

// NewConnector ...
func NewConnector(_ *testing.T) *Connector {
	mock := &Connector{ConnectorRecord: &ConnectorRecord{}}

	mock.ConnectFn = func(ctx context.Context, config *etcd.ConfigSpec) (etcd.Client, error) {
		return mock.Client, nil
	}

	mock.NewElectionFn = func(context.Context, string) (etcd.Election, error) {
		return mock.Election, nil
	}

	mock.ResumeElectionFn = func(context.Context, string, clientv3.GetResponse) (etcd.Election, error) {
		return mock.Election, nil
	}

	mock.StopFn = func() base.DoneChan {
		return mock.stopChan
	}

	mock.CleanupFn = func() {
	}

	return mock
}

// NewClient ...
func NewClient() *Client {
	watch := make(clientv3.WatchChan)
	client := Client{
		WatchFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
			return watch
		},
		GetFn: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return &clientv3.GetResponse{}, nil
		},
	}
	client.WatchCan = watch
	return &client
}

// NewElection ...
func NewElection() *Election {
	leaderChan := make(chan clientv3.GetResponse, 1)
	return &Election{
		leaderChan:     leaderChan,
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
	mock.stopChan = make(chan struct{}, 1)
	mock.Client = NewClient()
	mock.ConnectHits++
	return mock.ConnectFn(ctx, config)
}

// NewElection ...
func (mock *Connector) NewElection(ctx context.Context, electionKey string) (etcd.Election, error) {
	mock.Election = NewElection()
	mock.Election.ElectionKey = electionKey
	mock.NewElectionHits++
	return mock.NewElectionFn(ctx, electionKey)
}

// ResumeElection ...
func (mock *Connector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (etcd.Election, error) {
	mock.Election = NewElection()
	mock.Election.ElectionKey = electionKey
	mock.Election.Leader = leader
	mock.ResumeElectionHits++
	return mock.ResumeElectionFn(ctx, electionKey, leader)
}

// Stop ...
func (mock *Connector) Stop() base.DoneChan {
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
	close(mock.stopChan)
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
		close(mock.leaderChan)
	}()
	return mock.ObserveFn(ctx)
}

func (mock *Election) PushLeader(leader clientv3.GetResponse) {
	mock.leaderChan <- leader
	time.Sleep(10 * time.Millisecond)
}
