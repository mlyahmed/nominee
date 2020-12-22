package race

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type EtcdObserver struct {
	*Etcd
	proxy.Proxy
}

func NewEtcdObserver(config *EtcdConfig) Observer {
	logger = logrus.WithFields(logrus.Fields{"observer": "etcd"})
	return &EtcdObserver{
		Etcd: NewEtcd(config),
	}
}

func (observer *EtcdObserver) Observe(proxy proxy.Proxy) error {
	observer.Proxy = proxy
	observer.setUpOSSignals()
	if err := observer.subscribeToElection(); err != nil {
		return err
	}

	observer.stayTuned()
	observer.pushCurrentNominees()
	observer.observeLeaderNominee()
	observer.observeNominees()
	return nil
}

func (observer *EtcdObserver) observeNominees() {
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
					logger.Infof("Nominee deleted : %s", decoded.Marshal())
				case mvccpb.PUT:
					if err := observer.PushNominees(decoded); err != nil {
						panic(err)
					}
					logger.Infof("New Nominee added : %s", decoded.Marshal())
				default:
					panic("unknown")
				}
			}
		}
	}()
}

func (observer *EtcdObserver) observeLeaderNominee() {
	go func() {
		observe := observer.election.Observe(observer.ctx)
		for leader := range observe {
			decoded := observer.toNominee(leader)
			if err := observer.PushLeader(decoded); err != nil {
				panic(err)
			}
			logger.Infof("Leader pushed : %s", decoded.Marshal())
		}
	}()
}

func (observer *EtcdObserver) pushCurrentNominees() {
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
		logger.Infof("Nominee pushed : %s", decoded.Marshal())
	}
}

func (observer *EtcdObserver) subscribeToElection() error {
	logger.Infof("Subscribe to the election %s...", observer.electionKey())
	if err := observer.Etcd.newSession(); err != nil {
		return err
	}
	logger.Infof("new election...")
	observer.election = concurrency.NewElection(observer.session, observer.electionKey())
	logger.Infof("session created.")
	return nil
}
