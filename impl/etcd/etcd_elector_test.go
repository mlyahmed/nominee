package etcd_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/impl/etcd"
	etcdmock "github/mlyahmed.io/nominee/impl/mock"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"io/ioutil"
	"testing"
	"time"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

// The conform
func TestEtcdElector_must_be_conform(t *testing.T) {
	for _, example := range examples {
		t.Run("", func(t *testing.T) {
			election.TestElector(t, func() election.Elector {
				elector := etcd.NewElector(example.config)
				connector := etcdmock.NewConnector(t)
				elector.Connector = connector
				return elector
			})
		})
	}
}

func TestEtcdElector_when_run_then_connect_and_start_new_election(t *testing.T) {
	for _, example := range examples {
		t.Run(example.description, func(t *testing.T) {

			racer := etcd.NewElector(example.config)
			defer racer.Cleanup()
			connector := etcdmock.NewConnector(t)
			racer.Connector = connector

			if err := racer.Run(mock.NewNode(t, &node.Spec{})); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer, %v", testutils.Failed, err)
			}

			if connector.ConnectHits == 0 {
				t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to connect to the server. But actually not.", testutils.Failed)
			}

			if connector.NewElectionHits == 0 || connector.ResumeElectionHits != 0 {
				t.Fatalf("\t\t%s FAIL: EtcdRacer, expected a new election to start. But actually not.", testutils.Failed)
			}

		})
	}

}

func TestEtcdRacer_when_start_new_election_then_the_key_must_be_conform(t *testing.T) {

	for _, example := range examples {
		t.Run("", func(t *testing.T) {
			racer := etcd.NewElector(example.config)
			defer racer.Cleanup()
			connector := etcdmock.NewConnector(t)
			racer.Connector = connector

			if err := racer.Run(mock.NewNode(t, &node.Spec{})); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
			}

			expected := fmt.Sprintf("nominee/domain/%s/cluster/%s", example.config.Domain, example.config.Cluster)
			if expected != connector.Election.ElectionKey {
				t.Fatalf("\t\t%s FAIL: EtcdRacer, expected the election key <%s> but actual is <%s>", testutils.Failed, expected, connector.Election.ElectionKey)
			}
		})
	}

}

func TestEtcdRacer_must_campaign_for_leadership(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer, %v", testutils.Failed, err)
		}

		// Since in the contract, Etcd Campaign is a blocking function, it is invoked in a GOROUTINE. So we freeze a bit to let it be launched.
		time.Sleep(100 * time.Millisecond)
		if connector.Election.CampaignHits != 1 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to campaign for leadership. But actually not.", testutils.Failed)
		}

		expected := example.nominee.Marshal()
		actual := connector.Election.CampaignValue
		if expected != actual {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected campaign with value <%s> but actual is <%s>", testutils.Failed, expected, actual)
		}
	}
}

func TestEtcdRacer_when_elected_then_promote_the_service(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		srv.LeadFn = func(context.Context, node.Spec) error { return nil }
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		leader := example.toEtcdResponse()
		connector.Election.PushLeader(leader)

		if srv.LeadHits != 1 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to promote the node. But actually not.", testutils.Failed)
		}

		if string(leader.Kvs[0].Key) != srv.Leader.ElectionKey {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to promote the node itself. But actually not.", testutils.Failed)
		}

	}

}

func TestEtcdRacer_when_err_on_promote_then_stonith(t *testing.T) {
	for _, example := range examples {
		mockService := mock.NewNode(t, example.nominee)
		etcdRacer := etcd.NewElector(example.config)
		mockServerConnector := etcdmock.NewConnector(t)
		etcdRacer.Connector = mockServerConnector

		if err := etcdRacer.Run(mockService); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		mockService.LeadFn = func(context.Context, node.Spec) error {
			return errors.New("")
		}

		mockServerConnector.Election.PushLeader(example.toEtcdResponse())
		testutils.ItMustBeStopped(t, etcdRacer.Done())
	}
}

func TestEtcdRacer_when_another_nominee_is_promoted_then_follow_it(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, &node.Spec{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()})
		srv.FollowFn = func(context.Context, node.Spec) error { return nil }
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		leader := example.toEtcdResponse()
		connector.Election.PushLeader(leader)

		if srv.FollowHits != 1 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to follow. But actually not.", testutils.Failed)
		}

		if string(leader.Kvs[0].Key) != srv.Leader.ElectionKey {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to follow the promoted node. But actually not.", testutils.Failed)
		}
	}
}

func TestEtcdRacer_when_error_on_follow_then_stonith(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, &node.Spec{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()})
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		srv.FollowFn = func(context.Context, node.Spec) error {
			return errors.Errorf("")
		}

		connector.Election.PushLeader(example.toEtcdResponse())
		testutils.ItMustBeStopped(t, racer.Done())
	}
}

func TestEtcdRacer_when_another_leader_replaces_it_then_stonith(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector
		srv.LeadFn = func(ctx context.Context, nominee node.Spec) error { return nil }
		srv.StonithFn = func(ctx context.Context) error { return nil }

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		connector.Election.PushLeader(example.toEtcdResponse())

		n := &node.Spec{Name: string(uuid.NodeID()), Address: uuid.NodeInterface()}
		connector.Election.PushLeader(clientv3.GetResponse{
			Kvs: []*mvccpb.KeyValue{
				{
					Key:   uuid.NodeID(),
					Value: []byte(n.Marshal()),
				},
			},
		})

		if srv.StonithHits != 1 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected to stonith the node. But actually not.", testutils.Failed)
		}

		testutils.ItMustBeStopped(t, racer.Done())
	}
}

func TestEtcdRacer_when_the_server_session_is_closed_retry_to_connect(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer, error when RUN %v", testutils.Failed, err)
		}

		connector.CloseSession()

		if connector.ConnectHits != 2 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to retry to connect. But actually not.", testutils.Failed)
		}

		if connector.NewElectionHits != 2 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to create new election. But actually not.", testutils.Failed)
		}
	}
}

func TestEtcdRacer_when_there_was_already_a_leader_and_reconnect_then_resume_the_election(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		srv.LeadFn = func(context.Context, node.Spec) error { return nil }
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector

		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer, error when RUN %v", testutils.Failed, err)
		}

		leader := example.toEtcdResponse()
		connector.Election.PushLeader(leader)

		connector.CloseSession()

		if connector.ConnectHits != 2 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to retry to connect. But actually not.", testutils.Failed)
		}

		if connector.ResumeElectionHits != 1 && connector.NewElectionHits == 1 {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to resume the election. But actually not.", testutils.Failed)
		}

		if string(leader.Kvs[0].Key) != string(connector.Election.Leader.Kvs[0].Key) || leader.Kvs[0].CreateRevision != connector.Election.Leader.Kvs[0].CreateRevision {
			t.Fatalf("\t\t%s FAIL: EtcdRacer, expected  to resume the election with the same leader. But actually not.", testutils.Failed)
		}
	}

}

func TestEtcdRacer_when_service_is_stopped_then_stonith(t *testing.T) {
	for _, example := range examples {
		srv := mock.NewNode(t, example.nominee)
		racer := etcd.NewElector(example.config)
		connector := etcdmock.NewConnector(t)
		racer.Connector = connector
		if err := racer.Run(srv); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdRacer.Run, %v", testutils.Failed, err)
		}

		srv.StopChan <- struct{}{}

		time.Sleep(100 * time.Millisecond)
		testutils.ItMustBeStopped(t, racer.Done())
	}
}
