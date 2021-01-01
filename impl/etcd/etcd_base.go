package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

// Etcd ...
type Etcd struct {
	*stonither.Base
	*Config
	Connector    Connector
	nodeStopChan nominee.StopChan
	failBackFn   func() error
}

var (
	logger *logrus.Entry
)

// NewEtcd ...
func NewEtcd(config *Config) *Etcd {
	return &Etcd{
		Connector: NewDefaultConnector(),
		Base:      stonither.NewBase(),
		Config:    config,
	}
}

// Cleanup ...
func (etcd *Etcd) Cleanup() {
	etcd.Connector.Cleanup()
}

func (etcd *Etcd) listenToTheConnectorSession() {
	go func() {
		for {
			<-etcd.Connector.Stop()
			logger.Infof("session closed. Try to reconnect...")
			_ = etcd.failBackFn()
		}
	}()
}

func (etcd *Etcd) toNominee(response clientv3.GetResponse) nominee.NodeSpec {
	var value nominee.NodeSpec
	if len(response.Kvs) > 0 {
		value, _ = nominee.Unmarshal(response.Kvs[0].Value)
		value.ElectionKey = string(response.Kvs[0].Key)
	}
	return value
}

func (etcd *Etcd) electionKey() string {
	return fmt.Sprintf("nominee/domain/%s/cluster/%s", etcd.Domain, etcd.Cluster)
}
