package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"github/mlyahmed.io/nominee/pkg/signals"
	"os"
	"os/signal"
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

// ServerConnector ...
type ServerConnector interface {
	Connect(ctx context.Context, config *etcdconfig.Config) (Client, error)
	NewElection(ctx context.Context, electionKey string) (Election, error)
	ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error)
	Stop() <-chan struct{}
	Cleanup()
}

// DefaultServerConnector ...
type DefaultServerConnector struct {
	client  *clientv3.Client
	session *concurrency.Session
}

// Etcd ...
type Etcd struct {
	ServerConnector
	*etcdconfig.Config
	ctx             context.Context
	cancel          func()
	errorChan       chan error
	stopChan        chan struct{}
	stopped         bool
	nomineeStopChan nominee.StopChan
	failBackFn      func() error
}

var (
	logger *logrus.Entry
)

// NewEtcdConnectorServer ...
func NewEtcdConnectorServer() *DefaultServerConnector {
	return &DefaultServerConnector{}
}

// Connect ...
func (server *DefaultServerConnector) Connect(ctx context.Context, config *etcdconfig.Config) (Client, error) {
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
func (server *DefaultServerConnector) NewElection(_ context.Context, electionKey string) (Election, error) {
	election := concurrency.NewElection(server.session, electionKey)
	return election, nil
}

// ResumeElection ...
func (server *DefaultServerConnector) ResumeElection(ctx context.Context, electionKey string, leader clientv3.GetResponse) (Election, error) {
	election := concurrency.ResumeElection(server.session, electionKey, string(leader.Kvs[0].Key), leader.Kvs[0].CreateRevision)
	return election, nil
}

// StopChan ...
func (server *DefaultServerConnector) Stop() <-chan struct{} {
	return server.session.Done()
}

// Cleanup ...
func (server *DefaultServerConnector) Cleanup() {
	if server.client != nil {
		_ = server.client.Close()
	}

	if server.session != nil {
		_ = server.session.Close()
	}
}

// NewEtcd ...
func NewEtcd(config *etcdconfig.Config) *Etcd {
	ctx, cancel := context.WithCancel(context.Background())
	etcd := &Etcd{
		Config:          config,
		ctx:             ctx,
		cancel:          cancel,
		errorChan:       make(chan error),
		stopChan:        make(chan struct{}),
		ServerConnector: NewEtcdConnectorServer(),
	}

	return etcd
}

// Cleanup ...
func (etcd *Etcd) Cleanup() {
	etcd.ServerConnector.Cleanup()
}

// Stop ...
func (etcd *Etcd) Stop() nominee.StopChan {
	return etcd.stopChan
}

func (etcd *Etcd) setUpOSSignals() {
	listener := make(chan os.Signal, len(signals.ShutdownSignals))
	signal.Notify(listener, signals.ShutdownSignals...)
	go func() {
		<-listener
		etcd.stonith()
		<-listener
		os.Exit(1)
	}()
}

func (etcd *Etcd) setUpChannels() {
	go func() {
		for {
			select {
			case stop := <-etcd.nomineeStopChan:
				logger.Warningf("nominee stopped because of %s.", stop)
				etcd.stonith()
			case err := <-etcd.errorChan:
				if err != nil {
					etcd.stonith()
				}
			case <-etcd.ServerConnector.Stop():
				if etcd.stopped {
					return
				}
				logger.Infof("session closed. Try to reconnect...")
				_ = etcd.failBackFn()
			}
		}
	}()
}

func (etcd *Etcd) stonith() {
	etcd.cancel()
	close(etcd.stopChan)
	etcd.stopped = true
}

func (etcd *Etcd) toNominee(response clientv3.GetResponse) nominee.Nominee {
	var value nominee.Nominee
	if len(response.Kvs) > 0 {
		value, _ = nominee.Unmarshal(response.Kvs[0].Value)
		value.ElectionKey = string(response.Kvs[0].Key)
	}
	return value
}

func (etcd *Etcd) electionKey() string {
	return fmt.Sprintf("nominee/domain/%s/cluster/%s", etcd.Domain, etcd.Cluster)
}
