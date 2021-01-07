package etcd_test

import (
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/impl/etcd"
	etcdmock "github/mlyahmed.io/nominee/impl/mock"
	"github/mlyahmed.io/nominee/pkg/election"
	"io/ioutil"
	"testing"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestEtcdObserver_must_be_conform(t *testing.T) {
	for _, example := range examples {
		t.Run("", func(t *testing.T) {
			election.TestObserver(t, func() election.Observer {
				observer := etcd.NewObserver(example.config)
				observer.Connector = etcdmock.NewConnector(t)
				return observer
			})
		})
	}
}
