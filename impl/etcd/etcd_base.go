package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/node"
)

// Etcd ...
type Etcd struct {
	*ConfigSpec
	Connector  Connector
	failBackFn func() error
}

var (
	log *logrus.Entry
)

// NewEtcd ...
func NewEtcd(cl ConfigLoader) *Etcd {
	return &Etcd{
		Connector:  NewDefaultConnector(),
		ConfigSpec: cl.GetSpec(),
	}
}

// Cleanup ...
func (etcd *Etcd) Cleanup() {
	etcd.Connector.Cleanup()
}

func (etcd *Etcd) listenToTheConnectorSession() {
	go func() {
		for { //TODO: retries limit
			<-etcd.Connector.Stop()
			log.Infof("session closed. Try to reconnect...")
			_ = etcd.failBackFn()
		}
	}()
}

func (etcd *Etcd) toNodeSpec(response clientv3.GetResponse) node.Spec {
	var value node.Spec
	if len(response.Kvs) > 0 {
		value, _ = node.Unmarshal(response.Kvs[0].Value)
		value.ElectionKey = string(response.Kvs[0].Key)
	}
	return value
}

func (etcd *Etcd) electionKey() string {
	return fmt.Sprintf("nominee/domain/%s/cluster/%s", etcd.Domain, etcd.Cluster)
}
