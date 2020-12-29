package etcd_test

import (
	"context"
	"fmt"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/race/etcd"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"github/mlyahmed.io/nominee/pkg/service"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"testing"
)

func TestEtcdRacer_Election_Key(t *testing.T) {

	etcdRacerConfig := &etcdconfig.Config{
		BasicConfig: &config.BasicConfig{
			Cluster: "cluster-001",
			Domain:  "domain-001",
		},
		Endpoints: []string{"127.0.0.1:2154", "192.168.0.1:2379", "192.168.0.2:2379"},
	}

	etcdRacer := etcd.NewEtcdRacer(etcdRacerConfig)
	mockServerConnector := etcd.NewMockServerConnector()
	etcdRacer.ServerConnector = mockServerConnector
	mockServerConnector.NewElectionFn = func(ctx context.Context, electionKey string) (etcd.Election, error) {

		if electionKey != fmt.Sprintf("nominee/domain/%s/cluster/%s", etcdRacerConfig.Domain, etcdRacerConfig.Cluster) {
			t.Fatalf("\t\t%s FAIL: EtcdRacer.NewElection, expected <%s> but actual is <%s>", testutils.Failed, "", electionKey)
		}

		return mockServerConnector.Election, nil
	}

	_ = etcdRacer.Run(service.NewMockService())
}
