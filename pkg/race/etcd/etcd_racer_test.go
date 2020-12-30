package etcd_test

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
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
				mockService := service.NewMockServiceWithNominee(example.nominee)
				etcdRacer := etcd.NewEtcdRacer(example.config)
				mockServerConnector := etcd.NewMockServerConnector()
				etcdRacer.ServerConnector = mockServerConnector
				mockServerConnector.Election.CampaignFn = func(ctx context.Context, val string) error {
					conquering = true
					expected := example.nominee.Marshal()
					if example.nominee.Marshal() != val {
						t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected conquer with value <%s> but actual is <%s>", testutils.Failed, expected, val)
					}
					t.Logf("\t\t%s Must conquer with marshaled nominee as value.", testutils.Succeed)
					return nil
				}

				if err := etcdRacer.Run(mockService); err != nil {
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

func TestEtcdRacer_when_elected_then_promote_the_service(t *testing.T) {
	t.Logf("Given EtcdRacer is started.")
	{
		for i, example := range examples {
			promoted := false
			mockService := service.NewMockServiceWithNominee(example.nominee)
			mockService.LeadFn = func(context.Context, nominee.Nominee) error {
				promoted = true
				return nil
			}
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector
			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			t.Logf("\tTest %d: When it is promoted and %s", i, example.description)
			{
				mockServerConnector.Election.LeaderChan <- example.toEtcdResponse()
			}

			time.Sleep(10 * time.Millisecond)
			if !promoted {
				t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected to promote the service. But actually not.", testutils.Failed)
			}
			t.Logf("\t\t%s Then the service is promoted.", testutils.Succeed)
		}

	}
}

func TestEtcdRacer_when_service_is_stopped_then_stonith(t *testing.T) {
	t.Logf("Given EtcdRacer is running as the leader.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector
			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			t.Logf("\tTest %d: When the service is stopped and %s", i, example.description)
			{
				mockService.StopChan <- struct{}{}
				time.Sleep(100 * time.Millisecond)
				select {
				case <-etcdRacer.Stop():
					t.Logf("\t\t%s It must stonith.", testutils.Succeed)
				default:
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to stonith. But actually not.", testutils.Failed)
				}
			}

		}

	}
}
