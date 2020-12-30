package etcd_test

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcd"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"github/mlyahmed.io/nominee/pkg/service"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"io/ioutil"
	"testing"
	"time"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

type exampleSpec struct {
	description string
	config      *etcdconfig.Config
	nominee     *nominee.Nominee
}

var examples = []exampleSpec{
	{
		description: "one node cluster",
		config: &etcdconfig.Config{
			Endpoints:   []string{"127.0.0.1:2379"},
			BasicConfig: &config.BasicConfig{Cluster: "cluster-001", Domain: "domain-001"},
		},
		nominee: &nominee.Nominee{Name: "nominee-1", Address: "nominee-1", Port: 1245},
	},
	{
		description: "three nodes cluster",
		config: &etcdconfig.Config{
			Endpoints:   []string{"etcd1:2379", "etcd2:2379", "etcd3:2379"},
			BasicConfig: &config.BasicConfig{Cluster: "cluster-501", Domain: "domain-981"},
		},
		nominee: &nominee.Nominee{Name: "nominee-2", Address: "nominee-2", Port: 3254},
	},
	{
		description: "cluster with authentication",
		config: &etcdconfig.Config{
			Endpoints:   []string{"node1.etcd-cluster.priv", "node2.etcd-cluster.priv", "node3.etcd-cluster.priv"},
			Username:    "etcd-user",
			Password:    "21154)(*&^%@#_-_-_",
			BasicConfig: &config.BasicConfig{Cluster: "cluster-501", Domain: "domain-981"},
		},
		nominee: &nominee.Nominee{Name: "nominee-3", Address: "nominee-3", Port: 9778},
	},
}

func TestEtcdRacer_must_connect_and_start_new_election(t *testing.T) {
	t.Logf("Given EtcdRacer is stopped.")
	{
		for i, example := range examples {
			t.Logf("\tTest %d: When Run EtcdRacer and %s", i, example.description)
			{
				connected := false
				electionStarted := false
				etcdRacer := etcd.NewEtcdRacer(example.config)
				mockServerConnector := etcd.NewMockServerConnector()
				etcdRacer.ServerConnector = mockServerConnector

				mockServerConnector.ConnectFn = func(ctx context.Context, config *etcdconfig.Config) (etcd.Client, error) {
					connected = true
					return mockServerConnector.Client, nil
				}
				mockServerConnector.NewElectionFn = func(ctx context.Context, electionKey string) (etcd.Election, error) {
					electionStarted = true
					return mockServerConnector.Election, nil
				}

				if err := etcdRacer.Run(service.NewMockService()); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
				}

				if !connected {
					t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected to connect to the server. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s Is connected to the server.", testutils.Succeed)

				if !electionStarted {
					t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected a new election to start. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s A new election must be started.", testutils.Succeed)
			}
		}

	}
}

func TestEtcdRacer_the_election_key_must_be_conform(t *testing.T) {
	t.Logf("Given EtcdRacer is stopped.")
	{
		for i, example := range examples {
			t.Logf("\tTest %d: When Run EtcdRacer and %s", i, example.description)
			{
				etcdRacer := etcd.NewEtcdRacer(example.config)
				mockServerConnector := etcd.NewMockServerConnector()
				etcdRacer.ServerConnector = mockServerConnector
				mockServerConnector.NewElectionFn = func(ctx context.Context, electionKey string) (etcd.Election, error) {
					expected := fmt.Sprintf("nominee/domain/%s/cluster/%s", example.config.Domain, example.config.Cluster)
					if expected != electionKey {
						t.Fatalf("\t\t%s FAIL: EtcdRacer.NewElection, expected <%s> but actual is <%s>", testutils.Failed, expected, electionKey)
					}
					t.Logf("\t\t%s The election key must be conform.", testutils.Succeed)
					return mockServerConnector.Election, nil
				}
				if err := etcdRacer.Run(service.NewMockService()); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
				}
			}
		}

	}
}

func TestEtcdRacer_must_conquer_for_leadership(t *testing.T) {
	t.Logf("Given EtcdRacer is stopped.")
	{
		for i, example := range examples {
			t.Logf("\tTest %d: When Run EtcdRacer and %s", i, example.description)
			{
				conquering := false
				etcdRacer := etcd.NewEtcdRacer(example.config)
				mockServerConnector := etcd.NewMockServerConnector()
				etcdRacer.ServerConnector = mockServerConnector
				mockServerConnector.Election.CampaignFn = func(ctx context.Context, val string) error {
					conquering = true
					return nil
				}

				if err := etcdRacer.Run(service.NewMockService()); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
				}

				// Since in the contract, Etcd Campaign is a blocking function, it is invoked in a GOROUTINE. So we freeze a bit to let it be launched.
				time.Sleep(100 * time.Millisecond)
				if !conquering {
					t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected to conquer for leadership. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s Then it is conquering for leadership.", testutils.Succeed)
			}
		}

	}
}
