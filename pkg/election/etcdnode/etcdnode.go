package etcdnode

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/service"
	"github/mlyahmed.io/nominee/pkg/signals"
	"go.etcd.io/etcd/clientv3/concurrency"
	"os"
	"os/signal"
	"time"
)

type EtcdNode struct {
	service    service.Service
	endpoints  []string
	ctx        context.Context
	cancel     func()
	client     *clientv3.Client
	kv         clientv3.KV
	session    *concurrency.Session
	election   *concurrency.Election
	leaderEtcd clientv3.GetResponse
	errorCh    chan error
	stopCh     chan struct{}
	stopped    bool
}

var (
	logger *logrus.Entry
)

func NewEtcdNode(service service.Service, endpoints []string) *EtcdNode {
	subContext, subCancel := context.WithCancel(context.Background())
	logger = logrus.WithFields(logrus.Fields{"elector": "etcd", "service": service.ServiceName(), "node": service.NodeName()})
	return &EtcdNode{
		service:   service,
		endpoints: endpoints,
		ctx:       subContext,
		cancel:    subCancel,
		errorCh:   make(chan error),
		stopCh:    make(chan struct{}),
	}
}

func (node *EtcdNode) Cleanup() {
	if node.client != nil {
		_ = node.client.Close()
	}

	if node.session != nil {
		_ = node.session.Close()
	}
}

func (node *EtcdNode) Run() error {
	logger.Infof("starting...")

	node.setUpOSSignals()

	if err := node.newSession(); err != nil {
		return err
	}

	node.conquer()
	node.observe()
	node.stayTuned()

	logger.Infof("started.")
	return nil
}

func (node *EtcdNode) StopCh() <-chan struct{} {
	return node.stopCh
}

func (node *EtcdNode) setUpOSSignals() {
	listener := make(chan os.Signal, len(signals.ShutdownSignals))
	signal.Notify(listener, signals.ShutdownSignals...)
	go func() {
		<-listener
		node.stonith()
		<-listener
		os.Exit(1)
	}()
}

func (node *EtcdNode) newSession() error {
	logger.Infof("create new session...")
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   node.endpoints,
		DialTimeout: 1 * time.Second,
	})
	if err != nil {
		return err
	}
	node.client = client
	node.kv = clientv3.NewKV(node.client)

	session, err := concurrency.NewSession(node.client, concurrency.WithTTL(1), concurrency.WithContext(node.ctx))
	if err != nil {
		return err
	}

	node.session = session

	if len(node.leaderEtcd.Kvs) > 0 {
		logger.Infof("resume election...")
		node.election = concurrency.ResumeElection(node.session, node.etcdPrefix(), string(node.leaderEtcd.Kvs[0].Key), node.leaderEtcd.Kvs[0].CreateRevision)
	} else {
		logger.Infof("new election...")
		node.election = concurrency.NewElection(node.session, node.etcdPrefix())
	}

	logger.Infof("session created.")
	return nil
}

func (node *EtcdNode) stayTuned() {
	go func() {
		for {
			select {
			case err := <-node.service.StopChan():
				logger.Warningf("service stopped because of %s.", err)
				node.stonith()
			case err := <-node.errorCh:
				if err != nil {
					if errors.Cause(err) == context.Canceled {
						logger.Errorf("receive cancel error. Proceed to STOP !")
						node.stonith()
						return
					}
					logger.Warnf("ignored error %s", err)
				}
			case <-node.session.Done():
				if node.stopped {
					return
				}
				logger.Infof("session closed. Retrying...")
				node.retry()
			}
		}
	}()
}

func (node *EtcdNode) conquer() {
	go func() {
		logger.Infof("conquer as %v...", node.service.NodeName())
		node.errorCh <- node.election.Campaign(node.ctx, node.service.Nominee().Marshal())
	}()
}

func (node *EtcdNode) observe() {
	go func() {
		observe := node.election.Observe(node.ctx)
		for leader := range observe {
			node.changeLeader(leader)
		}
		logger.Debug("observation stopped.")
	}()
}

func (node *EtcdNode) changeLeader(leader clientv3.GetResponse) {
	amICurrentlyTheLeader := node.amITheLeader()
	amITheNewLeader := node.toNominee(leader).Name == node.service.NodeName()
	node.leaderEtcd = leader

	if amITheNewLeader && amICurrentlyTheLeader {

		logger.Infof("I stay the leaderEtcd. Nothing to do.")

	} else if amITheNewLeader && !amICurrentlyTheLeader {

		logger.Infof("promoting The Service...")
		node.errorCh <- node.service.Promote(node.ctx, node.leaderNominee())

	} else if !amITheNewLeader && amICurrentlyTheLeader {

		node.stonith()

	} else {

		node.errorCh <- node.service.FollowNewLeader(node.ctx, node.leaderNominee())

	}
}

func (node *EtcdNode) retry() {
	logger.Infof("retrying...")

	node.cancel()
	node.ctx, node.cancel = context.WithCancel(context.Background())

	for {
		if err := node.newSession(); err != nil {
			logger.Errorf("error when retry new session, %s", err)
			time.Sleep(time.Second * 2)
		} else {
			logger.Info("new session created.")
			break
		}
	}

	node.conquer()
	node.observe()
}

func (node *EtcdNode) stonith() {
	logger.Infof("stonithing...")

	if node.amITheLeader() {
		logger.Infof("resign since I was leader...")
		_ = node.service.Stonith(node.ctx)
		_ = node.election.Resign(node.ctx)
	}

	node.cancel()
	close(node.stopCh)
	node.stopped = true
}

func (node *EtcdNode) leaderNominee() nominee.Nominee {
	return node.toNominee(node.leaderEtcd)
}

func (node *EtcdNode) amITheLeader() bool {
	return node.leaderNominee().Name == node.service.NodeName()
}

func (node *EtcdNode) toNominee(response clientv3.GetResponse) nominee.Nominee {
	var value nominee.Nominee
	if len(response.Kvs) > 0 {
		value, _ = nominee.Unmarshal(response.Kvs[0].Value)
	}
	return value
}

func (node *EtcdNode) etcdPrefix() string {
	return fmt.Sprintf("nominee/service/%s/cluster/%s", node.service.ServiceName(), node.service.ClusterName())
}
