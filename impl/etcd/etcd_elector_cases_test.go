package etcd_test

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/mock"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/node"
	"time"
)

type exampleSpec struct {
	description string
	config      *mock.ConfigSpec
	nominee     *node.Spec
}

func (example exampleSpec) toEtcdResponse() clientv3.GetResponse {
	return clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Key:            uuid.NodeID(),
				Value:          []byte(example.nominee.Marshal()),
				CreateRevision: time.Now().Unix(),
			},
		},
	}
}

var examples = []exampleSpec{
	{
		description: "one node cluster",
		config: &mock.ConfigSpec{
			ConfigSpec: &etcd.ConfigSpec{
				Endpoints:   []string{"127.0.0.1:2379"},
				BasicConfig: &config.BasicConfig{Cluster: "cluster-001", Domain: "domain-001"},
			},
		},
		nominee: &node.Spec{Name: "nominee-1", Address: "nominee-1", Port: 1245},
	},
	{
		description: "three nodes cluster",
		config: &mock.ConfigSpec{
			ConfigSpec: &etcd.ConfigSpec{
				Endpoints:   []string{"etcd1:2379", "etcd2:2379", "etcd3:2379"},
				BasicConfig: &config.BasicConfig{Cluster: "cluster-501", Domain: "domain-981"},
			},
		},
		nominee: &node.Spec{Name: "nominee-2", Address: "nominee-2", Port: 3254},
	},
	{
		description: "cluster with authentication",
		config: &mock.ConfigSpec{
			ConfigSpec: &etcd.ConfigSpec{
				Endpoints:   []string{"node1.etcd-cluster.priv", "node2.etcd-cluster.priv", "node3.etcd-cluster.priv"},
				Username:    "etcd-user",
				Password:    "21154)(*&^%@#_-_-_",
				BasicConfig: &config.BasicConfig{Cluster: "cluster-501", Domain: "domain-981"},
			},
		},
		nominee: &node.Spec{Name: "nominee-3", Address: "nominee-3", Port: 9778},
	},
}
