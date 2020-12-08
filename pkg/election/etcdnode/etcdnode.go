package etcdnode

import (
	"context"
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
	service   service.Service
	endpoints []string
	ctx       context.Context
	cancel    func()
	client    *clientv3.Client
	session   *concurrency.Session
	election  *concurrency.Election
	leader    clientv3.GetResponse
	errorCh   chan error
	stopCh    chan struct{}
	stopped   bool
}

var (
	logger *logrus.Entry
)

func NewEtcdNode(service service.Service, endpoints []string) *EtcdNode {
	subContext, subCancel := context.WithCancel(context.Background())
	logger = logrus.WithFields(logrus.Fields{
		"elector": "etcd",
		"service": service.Name(),
		"node": service.NodeName(),
	})

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
	logger.Infof("etcd: starting...")

	node.setUpSignals()

	if err := node.newSession(); err != nil {
		return err
	}

	node.conquer()
	node.observe()
	node.stayTuned()

	logger.Infof("etcd: Started.")
	return nil
}

func (node *EtcdNode) StopCh() <-chan struct{} {
	return node.stopCh
}

func (node *EtcdNode) setUpSignals() {
	listener := make(chan os.Signal, len(signals.ShutdownSignals))
	signal.Notify(listener, signals.ShutdownSignals...)
	go func() {
		<-listener
		node.stop()
		<-listener
		os.Exit(1)
	}()
}

func (node *EtcdNode) newSession() error {
	logger.Infof("etcd: create new session...")
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   node.endpoints,
		DialTimeout: 1 * time.Second,
	})
	if err != nil {
		return err
	}
	node.client = client

	session, err := concurrency.NewSession(node.client, concurrency.WithTTL(1), concurrency.WithContext(node.ctx))
	if err != nil {
		return err
	}

	node.session = session
	if len(node.leader.Kvs) > 0 {
		logger.Infof("etcd: resume election...")
		node.election = concurrency.ResumeElection(node.session, node.service.ClusterName(), string(node.leader.Kvs[0].Key), node.leader.Kvs[0].CreateRevision)
	} else {
		logger.Infof("etcd: new election...")
		node.election = concurrency.NewElection(node.session, node.service.ClusterName())
	}
	logger.Infof("etcd: session created.")
	return nil
}

func (node *EtcdNode) stayTuned() {
	go func() {
		for {
			select {
			case err := <-node.service.StopChan():
				if err == nil {
					logger.Infof("etcd: service stopped.")
				} else {
					logger.Errorf("etcd: service stopped because of %s.", err)
				}
				node.stop()
			case err := <-node.errorCh:
				if err != nil {
					if errors.Cause(err) == context.Canceled {
						logger.Errorf("receive cancel error. Proceed to STOP !")
						node.stop()
						return
					}
					logger.Warnf("etcd: ignored error %s", err)
				}
			case <-node.session.Done():
				if node.stopped {
					return
				}
				logger.Infof("etcd: session closed. Retrying...")
				node.retry()
			}
		}
	}()
}

func (node *EtcdNode) conquer() {
	go func() {
		logger.Infof("etcd: conquer as %v...", node.service.NodeName())
		node.errorCh <- node.election.Campaign(node.ctx, node.service.NodeName())
	}()
}

func (node *EtcdNode) observe() {
	go func() {
		observe := node.election.Observe(node.ctx)
		for leader := range observe {
			node.changeLeader(leader)
		}
		logger.Debug("etcd: observation stopped.")
	}()
}

func (node *EtcdNode) changeLeader(leader clientv3.GetResponse) {
	amICurrentlyTheLeader := node.amITheLeader()
	node.leader = leader
	amITheNewLeader := string(leader.Kvs[0].Value) == node.service.NodeName()

	if amITheNewLeader && amICurrentlyTheLeader {

		logger.Infof("etcd: I stay the leader. Nothing to do.")

	} else if amITheNewLeader && !amICurrentlyTheLeader {
		logger.Infof("etcd: Promoting The Service...")
		if err := node.service.Promote(node.ctx, nominee.Nominee{}); err != nil {
			node.errorCh <- err
			return
		}
	} else if !amITheNewLeader && amICurrentlyTheLeader {
		//STONITH : Shoot The EtcdNode In The Head
		logger.Infof("etcd: shoot me in the head...")
		node.stop()

	} else {
		node.errorCh <- node.service.FollowNewLeader(node.ctx, nominee.Nominee{})
	}
}

func (node *EtcdNode) retry() {
	logger.Infof("etcd: retrying...")

	node.cancel()
	node.ctx, node.cancel = context.WithCancel(context.Background())

	for {
		if err := node.newSession(); err != nil {
			logger.Errorf("etcd: error when retry new session, %s", err)
			time.Sleep(time.Second * 2)
		} else {
			logger.Info("etcd: new session created.")
			break
		}
	}

	node.conquer()
	node.observe()
}

func (node *EtcdNode) stop() {
	logger.Infof("etcd: Stopping...")

	if node.amITheLeader() {
		logger.Infof("etcd: resign since I was leader...")
		_ = node.service.Stonith(node.ctx)
		_ = node.election.Resign(node.ctx)
	}

	node.cancel()
	close(node.stopCh)
	node.stopped = true
}

func (node *EtcdNode) amITheLeader() bool {
	return len(node.leader.Kvs) > 0 && string(node.leader.Kvs[0].Value) == node.service.NodeName()
}
