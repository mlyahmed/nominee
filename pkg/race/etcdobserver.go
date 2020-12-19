package race

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type EtcdObserver struct {
	*Etcd
	*proxy.Config
	proxy.Proxy
}

func NewEtcdObserver(endpoints []string) Observer {
	subContext, subCancel := context.WithCancel(context.Background())
	logger = logrus.WithFields(logrus.Fields{"observer": "etcd"})
	return &EtcdObserver{
		Etcd: &Etcd{
			endpoints: endpoints,
			ctx:       subContext,
			cancel:    subCancel,
			errorChan: make(chan error),
			stopChan:  make(chan error),
		},
	}
}

func (observer *EtcdObserver) Observe(proxy proxy.Proxy) error {
	observer.Proxy = proxy
	observer.domain = proxy.Config().Domain // Very bad
	observer.cluster = proxy.Config().Cluster

	observer.setUpOSSignals()
	if err := observer.subscribeToElection(); err != nil {
		panic(err)
	}

	response, err := observer.client.Get(observer.ctx, observer.electionKey(), clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	for _, n := range response.Kvs {
		decoded, _ := nominee.Unmarshal(n.Value)
		decoded.ElectionKey = string(n.Key)
		if err := observer.PushNominees(decoded); err != nil {
			panic(err)
		}
	}

	go func() {
		observe := observer.election.Observe(observer.ctx)
		for leader := range observe {
			if err := observer.PushLeader(observer.toNominee(leader)); err != nil {
				panic(err)
			}
		}
	}()

	go func() {
		watch := observer.client.Watch(observer.ctx, observer.electionKey(), clientv3.WithPrefix())
		for response := range watch {
			for _, v := range response.Events {
				decoded, _ := nominee.Unmarshal(v.Kv.Value)
				decoded.ElectionKey = string(v.Kv.Key)
				switch v.Type {
				case mvccpb.DELETE:
					if err := observer.RemoveNominee(string(v.Kv.Key)); err != nil {
						panic(err)
					}
				case mvccpb.PUT:
					if err := observer.PushNominees(decoded); err != nil {
						panic(err)
					}
				default:
					panic("unknown")
				}
			}
		}
	}()

	observer.stayTuned()
	return nil
}

func (observer *EtcdObserver) subscribeToElection() error {
	if err := observer.Etcd.newSession(); err != nil {
		return err
	}
	logger.Infof("new election...")
	observer.election = concurrency.NewElection(observer.session, observer.electionKey())
	logger.Infof("session created.")
	return nil
}
