package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/node"
	proxy2 "github/mlyahmed.io/nominee/pkg/proxy"
)

// Observer ...
type Observer struct {
	*Etcd
	proxy    proxy2.Proxy
	client   Client
	election Election
}

// NewEtcdObserver ...
func NewEtcdObserver(cl ConfigLoader) election.Observer {
	cl.Load(context.Background())
	log = logrus.WithFields(logrus.Fields{"observer": "etcd"})
	return &Observer{Etcd: NewEtcd(cl)}
}

// Observe ...
func (observer *Observer) Observe(proxy proxy2.Proxy) error {
	observer.proxy = proxy
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
		watch := observer.client.Watch(observer.Ctx, observer.electionKey(), clientv3.WithPrefix())
		for response := range watch {
			for _, v := range response.Events {
				decoded, _ := node.Unmarshal(v.Kv.Value)
				decoded.ElectionKey = string(v.Kv.Key)
				switch v.Type {
				case mvccpb.DELETE:
					if err := observer.proxy.RemoveNode(string(v.Kv.Key)); err != nil {
						panic(err)
					}
					log.Infof("GetSpec deleted : %s", decoded.Marshal())
				case mvccpb.PUT:
					if err := observer.proxy.PushNodes(decoded); err != nil {
						panic(err)
					}
					log.Infof("New GetSpec added : %s", decoded.Marshal())
				default:
					panic("unknown")
				}
			}
		}
	}()
}

func (observer *Observer) observeLeaderNominee() {
	go func() {
		observe := observer.election.Observe(observer.Ctx)
		for leader := range observe {
			decoded := observer.toNominee(leader)
			if err := observer.proxy.PushLeader(decoded); err != nil {
				panic(err)
			}
			log.Infof("Leader pushed : %s", decoded.Marshal())
		}
	}()
}

func (observer *Observer) pushCurrentNominees() {
	response, err := observer.client.Get(observer.Ctx, observer.electionKey(), clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	for _, n := range response.Kvs {
		decoded, _ := node.Unmarshal(n.Value)
		decoded.ElectionKey = string(n.Key)
		if err := observer.proxy.PushNodes(decoded); err != nil {
			panic(err)
		}
		log.Infof("GetSpec pushed : %s", decoded.Marshal())
	}
}

func (observer *Observer) subscribe() error {
	log.Infof("Subscribe to the election %s...", observer.electionKey())
	var err error
	if observer.client, err = observer.Connector.Connect(observer.Ctx, observer.ConfigSpec); err != nil {
		return err
	}
	log.Infof("new election...")
	observer.election, _ = observer.Connector.NewElection(observer.Ctx, observer.electionKey())

	log.Infof("subscribed to Etcd server.")
	return nil
}
