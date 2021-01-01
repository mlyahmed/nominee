package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

// Observer ...
type Observer struct {
	*Etcd
	nominee.Proxy
	client   Client
	election Election
}

// NewEtcdObserver ...
func NewEtcdObserver(config *Config) election.Observer {
	logger = logrus.WithFields(logrus.Fields{"observer": "etcd"})
	return &Observer{Etcd: NewEtcd(config)}
}

// Observe ...
func (observer *Observer) Observe(proxy nominee.Proxy) error {
	observer.Proxy = proxy
	if err := observer.subscribe(); err != nil {
		return err
	}

	observer.listenToTheConnectorSession()
	observer.pushCurrentNominees()
	observer.observeLeaderNominee()
	observer.observeNominees()
	return nil
}

func (observer *Observer) observeNominees() {
	go func() {
		watch := observer.client.Watch(observer.Context, observer.electionKey(), clientv3.WithPrefix())
		for response := range watch {
			for _, v := range response.Events {
				decoded, _ := nominee.Unmarshal(v.Kv.Value)
				decoded.ElectionKey = string(v.Kv.Key)
				switch v.Type {
				case mvccpb.DELETE:
					if err := observer.RemoveNode(string(v.Kv.Key)); err != nil {
						panic(err)
					}
					logger.Infof("NodeSpec deleted : %s", decoded.Marshal())
				case mvccpb.PUT:
					if err := observer.PushNodes(decoded); err != nil {
						panic(err)
					}
					logger.Infof("New NodeSpec added : %s", decoded.Marshal())
				default:
					panic("unknown")
				}
			}
		}
	}()
}

func (observer *Observer) observeLeaderNominee() {
	go func() {
		observe := observer.election.Observe(observer.Context)
		for leader := range observe {
			decoded := observer.toNominee(leader)
			if err := observer.PushLeader(decoded); err != nil {
				panic(err)
			}
			logger.Infof("Leader pushed : %s", decoded.Marshal())
		}
	}()
}

func (observer *Observer) pushCurrentNominees() {
	response, err := observer.client.Get(observer.Context, observer.electionKey(), clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	for _, n := range response.Kvs {
		decoded, _ := nominee.Unmarshal(n.Value)
		decoded.ElectionKey = string(n.Key)
		if err := observer.PushNodes(decoded); err != nil {
			panic(err)
		}
		logger.Infof("NodeSpec pushed : %s", decoded.Marshal())
	}
}

func (observer *Observer) subscribe() error {
	logger.Infof("Subscribe to the election %s...", observer.electionKey())
	var err error
	if observer.client, err = observer.Connector.Connect(observer.Context, observer.Config); err != nil {
		return err
	}
	logger.Infof("new election...")
	observer.election, _ = observer.Connector.NewElection(observer.Context, observer.electionKey())

	logger.Infof("subscribed to Etcd server.")
	return nil
}
