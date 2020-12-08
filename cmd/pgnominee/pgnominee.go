package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/election/etcdnode"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/service/postgres"
	"os"
	"strings"
	"time"
)

var (
	nodeName string
	nodeAddress string
	clusterName string
	etcdEndPoints string
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	var ok bool
	if nodeName, ok = os.LookupEnv("PG_NOMINEE_NODE_NAME"); !ok {
		logrus.Fatalf("You must specify the env var PG_NOMINEE_NODE_NAME to a non-empty value.")
	}

	if nodeAddress, ok = os.LookupEnv("PG_NOMINEE_NODE_ADDRESS"); !ok {
		logrus.Fatalf("You must specify the env var PG_NOMINEE_NODE_ADDRESS to a non-empty value (IP or DNS).")
	}

	if clusterName, ok = os.LookupEnv("PG_NOMINEE_CLUSTER_NAME"); !ok {
		logrus.Fatalf("You must specify the env var PG_NOMINEE_CLUSTER_NAME to a non-empty value.")
	}

	if etcdEndPoints, ok = os.LookupEnv("PG_NOMINEE_ETCD_ENDPOINTS"); !ok {
		logrus.Fatalf("You must specify the env var PG_NOMINEE_ETCD_ENDPOINTS to a non-empty value.")
	}
}



func main() {
	node := nominee.Nominee{
		Name: fmt.Sprintf("%s-%d", nodeName, time.Now().Nanosecond()), //Make sure the nodes do not collide. It is the pgnominee.main responsibility ?
		Cluster: clusterName,
		Address: nodeAddress,
	}
	service := postgres.NewPostgres(node)
	elector := etcdnode.NewEtcdNode(service, strings.Split(etcdEndPoints, ","))
	defer elector.Cleanup()
	if err := elector.Run(); err != nil {
		logrus.Errorf("pgnominee: %v \n", err)
		return
	}
	<-elector.StopCh()

	logrus.Infof("pgnominee: stopped.")
}


