package etcd_test

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/impl/etcd"
	etcdmock "github/mlyahmed.io/nominee/impl/mock"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"io/ioutil"
	"testing"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestEtcdObserver_must_be_conform(t *testing.T) {
	for _, example := range examples {
		t.Run(example.description, func(t *testing.T) {
			election.TestObserver(t, func() election.Observer {
				observer := etcd.NewObserver(example.config)
				observer.Connector = etcdmock.NewConnector(t)
				return observer
			})
		})
	}
}

func TestEtcdObserver_when_observe_then_connect_and_start_new_election(t *testing.T) {
	for _, example := range examples {
		t.Run(example.description, func(t *testing.T) {
			observer := etcd.NewObserver(example.config)
			defer observer.Cleanup()
			connector := etcdmock.NewConnector(t)
			observer.Connector = connector

			if err := observer.Observe(mock.NewProxy()); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdObserver, %v", testutils.Failed, err)
			}

			if connector.ConnectHits == 0 {
				t.Fatalf("\t\t%s FAIL: EtcdObserver, expected to connect to the server. But actually not.", testutils.Failed)
			}

			if connector.NewElectionHits == 0 || connector.ResumeElectionHits != 0 {
				t.Fatalf("\t\t%s FAIL: EtcdObserver, expected a new election to start. But actually not.", testutils.Failed)
			}
		})
	}
}

func TestEtcdObserver_when_subscribe_to_the_election_then_the_key_must_be_conform(t *testing.T) {
	for _, example := range examples {
		t.Run("", func(t *testing.T) {
			observer := etcd.NewObserver(example.config)
			defer observer.Cleanup()
			connector := etcdmock.NewConnector(t)
			observer.Connector = connector

			if err := observer.Observe(mock.NewProxy()); err != nil {
				t.Fatalf("\t\t%s FATAL: EtcdObserver.Observe, %v", testutils.Failed, err)
			}

			expected := fmt.Sprintf("nominee/domain/%s/cluster/%s", example.config.Domain, example.config.Cluster)
			if expected != connector.Election.ElectionKey {
				t.Fatalf("\t\t%s FAIL: EtcdObserver, expected the subscribe to the election <%s> but actual is <%s>", testutils.Failed, expected, connector.Election.ElectionKey)
			}
		})
	}
}

func TestEtcdObserver_when_the_server_session_is_closed_retry_to_connect(t *testing.T) {
	for _, example := range examples {
		observer := etcd.NewObserver(example.config)
		connector := etcdmock.NewConnector(t)
		observer.Connector = connector

		if err := observer.Observe(mock.NewProxy()); err != nil {
			t.Fatalf("\t\t%s FATAL: EtcdObserver, error when RUN %v", testutils.Failed, err)
		}

		connector.CloseSession()

		if connector.ConnectHits != 2 {
			t.Fatalf("\t\t%s FAIL: EtcdObserver, expected  to retry to connect. But actually not.", testutils.Failed)
		}

		if connector.NewElectionHits != 2 {
			t.Fatalf("\t\t%s FAIL: EtcdObserver, expected  to create new election. But actually not.", testutils.Failed)
		}
	}
}
