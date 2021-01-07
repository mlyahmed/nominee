package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
)

// Observer ...
type Observer struct {
	*Etcd
	*election.BasicObserver
	client   Client
	election Election
}

// NewObserver ...
func NewObserver(cl ConfigLoader) *Observer {
	cl.Load(context.Background())
	log = logrus.WithFields(logrus.Fields{"observer": "etcd"})
	return &Observer{Etcd: NewEtcd(cl)}
}

// Observe ...
func (observer *Observer) Observe(proxy proxy.Proxy) error {
	observer.BasicObserver = election.NewBasicObserver(proxy)

	if err := observer.subscribe(); err != nil {
		return err
	}

	observer.listenToTheConnectorSession()
	observer.pushCurrentNodes()
	observer.observeLeader()
	observer.observeNodes()
	return nil
}

func (observer *Observer) observeNodes() {
	go func() {
		watch := observer.client.Watch(observer.Ctx, observer.electionKey(), clientv3.WithPrefix())
		for response := range watch {
			for _, v := range response.Events {
				decoded, _ := node.Unmarshal(v.Kv.Value)
				decoded.ElectionKey = string(v.Kv.Key)
				switch v.Type {
				case mvccpb.DELETE:
					if err := observer.RemoveNodes(&decoded); err != nil {
						panic(err)
					}
					log.Infof("Node deleted : %s", decoded.Marshal())
				case mvccpb.PUT:
					if err := observer.UpdateNodes([]*node.Spec{&decoded}); err != nil {
						panic(err)
					}
					log.Infof("node updated : %s", decoded.Marshal())
				default:
					panic("unknown")
				}
			}
		}
	}()
}

func (observer *Observer) observeLeader() {
	go func() {
		observe := observer.election.Observe(observer.Ctx)
		for leader := range observe {
			decoded := observer.toNodeSpec(leader)
			if err := observer.UpdateLeader(&decoded); err != nil {
				panic(err)
			}
			log.Infof("Leader updated : %s", decoded.Marshal())
		}
	}()
}

func (observer *Observer) pushCurrentNodes() {
	response, err := observer.client.Get(observer.Ctx, observer.electionKey(), clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	nodes := make([]*node.Spec, len(response.Kvs))
	for i, n := range response.Kvs {
		decoded, _ := node.Unmarshal(n.Value)
		decoded.ElectionKey = string(n.Key)
		nodes[i] = &decoded
	}

	if err := observer.UpdateNodes(nodes); err != nil {
		panic(err)
	}
	log.Infof("Current nodes updated: %v", nodes)

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
