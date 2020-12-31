package etcd_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcd"
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
			etcdRacer := etcd.NewEtcdRacer(example.config)
			serverConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = serverConnector

			t.Logf("\tTest %d: When Run and %s", i, example.description)
			{
				if err := etcdRacer.Run(service.NewMockService(t)); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
				}

				if serverConnector.Statistics.ConnectHits == 0 {
					t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected to connect to the server. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s Is connected to the server.", testutils.Succeed)

				if serverConnector.Statistics.NewElectionHits == 0 {
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
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector
			actualKeyElection := ""
			mockServerConnector.NewElectionFn = func(ctx context.Context, electionKey string) (etcd.Election, error) {
				actualKeyElection = electionKey
				return mockServerConnector.Election, nil
			}

			t.Logf("\tTest %d: When Run and %s", i, example.description)
			{
				if err := etcdRacer.Run(service.NewMockService(t)); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
				}

				expected := fmt.Sprintf("nominee/domain/%s/cluster/%s", example.config.Domain, example.config.Cluster)
				if expected != actualKeyElection {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected the election key <%s> but actual is <%s>", testutils.Failed, expected, actualKeyElection)
				}
				t.Logf("\t\t%s The election key must be conform.", testutils.Succeed)
			}
		}

	}
}

func TestEtcdRacer_must_conquer_for_leadership(t *testing.T) {
	t.Logf("Given EtcdRacer is stopped.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector

			t.Logf("\tTest %d: When Run and %s", i, example.description)
			{
				if err := etcdRacer.Run(mockService); err != nil {
					t.Fatalf("\t\t%s FATAL: EtcdRacer, %v", testutils.Failed, err)
				}

				conquering := false
				mockServerConnector.Election.CampaignFn = func(ctx context.Context, val string) error {
					conquering = true
					expected := example.nominee.Marshal()
					if example.nominee.Marshal() != val {
						t.Fatalf("\t\t%s FAIL: EtcdRacer, expected conquer with value <%s> but actual is <%s>", testutils.Failed, expected, val)
					}
					t.Logf("\t\t%s Must conquer with marshaled nominee as value.", testutils.Succeed)
					return nil
				}

				// Since in the contract, Etcd Campaign is a blocking function, it is invoked in a GOROUTINE. So we freeze a bit to let it be launched.
				time.Sleep(100 * time.Millisecond)
				if !conquering {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to conquer for leadership. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s It is conquering for leadership.", testutils.Succeed)
			}
		}

	}
}

func TestEtcdRacer_when_elected_then_promote_the_service(t *testing.T) {
	t.Logf("Given EtcdRacer is started.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector

			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			promoted := false
			mockService.LeadFn = func(context.Context, nominee.Nominee) error {
				promoted = true
				return nil
			}

			t.Logf("\tTest %d: When it is promoted and %s", i, example.description)
			{
				mockServerConnector.Election.PushLeader(example.toEtcdResponse())

				if !promoted {
					t.Fatalf("\t\t%s FAIL: EtcdRacer.Run, expected to promote the service. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s Then the service is promoted.", testutils.Succeed)
			}

		}

	}
}

func TestEtcdRacer_when_err_on_promote_then_stonith(t *testing.T) {
	t.Logf("Given the service is bugged.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector

			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			mockService.LeadFn = func(context.Context, nominee.Nominee) error {
				return errors.New("")
			}

			t.Logf("\tTest %d: When it is promoted and %s", i, example.description)
			{
				mockServerConnector.Election.LeaderChan <- example.toEtcdResponse()

				time.Sleep(10 * time.Millisecond)
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

func TestEtcdRacer_when_another_nominee_is_promoted_then_follow_it(t *testing.T) {
	t.Logf("Given Etcd Racer is started.")
	{
		for i, example := range examples {
			follow := false
			mockService := service.NewMockServiceWithNominee(t, &nominee.Nominee{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()})
			mockService.FollowFn = func(context.Context, nominee.Nominee) error {
				follow = true
				return nil
			}
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector
			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			t.Logf("\tTest %d: When another nominee is promoted and %s", i, example.description)
			{
				mockServerConnector.Election.PushLeader(example.toEtcdResponse())

				time.Sleep(10 * time.Millisecond)
				if !follow {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to follow the new leader. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s Then the service follows the new leader.", testutils.Succeed)
			}

		}

	}
}

func TestEtcdRacer_when_error_on_follow_then_stonith(t *testing.T) {
	t.Logf("Given the Nominee is bugged.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, &nominee.Nominee{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()})
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector

			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			mockService.FollowFn = func(context.Context, nominee.Nominee) error {
				return errors.Errorf("")
			}

			t.Logf("\tTest %d: When another nominee is promoted and %s", i, example.description)
			{
				mockServerConnector.Election.LeaderChan <- example.toEtcdResponse()

				time.Sleep(10 * time.Millisecond)
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

func TestEtcdRacer_when_demoted_then_stonith(t *testing.T) {
	t.Logf("Given EtcdRacer is the leader.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			mockServerConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = mockServerConnector
			serviceStonithed := false
			mockService.LeadFn = func(ctx context.Context, nominee nominee.Nominee) error { return nil }
			mockService.StonithFn = func(ctx context.Context) error { serviceStonithed = true; return nil }

			if err := etcdRacer.Run(mockService); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			mockServerConnector.Election.PushLeader(example.toEtcdResponse())

			t.Logf("\tTest %d: When it is demoted and %s", i, example.description)
			{
				n := &nominee.Nominee{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()}
				mockServerConnector.Election.PushLeader(clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{
							Key:   uuid.NodeID(),
							Value: []byte(n.Marshal()),
						},
					},
				})

				if !serviceStonithed {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to stonith the service. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s The service must be stonith.", testutils.Succeed)

				select {
				case <-etcdRacer.Stop():
					t.Logf("\t\t%s EtcdRacer must stonith.", testutils.Succeed)
				default:
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to stonith. But actually not.", testutils.Failed)
				}
			}

		}

	}
}

func TestEtcdRacer_when_the_server_session_is_closed_retry_to_connect(t *testing.T) {
	t.Logf("Given EtcdRacer is running.")
	{
		for i, example := range examples {
			srv := service.NewMockServiceWithNominee(t, example.nominee)
			etcdRacer := etcd.NewEtcdRacer(example.config)
			serverConnector := etcd.NewMockServerConnector()
			etcdRacer.ServerConnector = serverConnector

			if err := etcdRacer.Run(srv); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer, error when RUN %v", testutils.Failed, err)
			}

			t.Logf("\tTest %d: When the server session is closed and %s", i, example.description)
			{
				serverConnector.CloseSession()

				//<- etcdRacer.Stop()
				if serverConnector.Statistics.ConnectHits < 2 {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to retry to connect. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s It must retry to connect.", testutils.Succeed)

				if serverConnector.Statistics.NewElectionHits < 2 {
					t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to create new election. But actually not.", testutils.Failed)
				}
				t.Logf("\t\t%s It must create new election.", testutils.Succeed)
			}

		}

	}
}

func TestEtcdRacer_when_service_is_stopped_then_stonith(t *testing.T) {
	t.Logf("Given EtcdRacer is running.")
	{
		for i, example := range examples {
			mockService := service.NewMockServiceWithNominee(t, example.nominee)
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
