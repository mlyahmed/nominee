package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/build"
	"github/mlyahmed.io/nominee/pkg/election/etcdnode"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/service/postgres"
	"os"
	"strings"
	"time"
)

var (
	nodeName         string
	nodeAddress      string
	clusterName      string
	etcdEndPoints    string
	postgresPassword string
	logger           *logrus.Entry
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logger = logrus.WithFields(logrus.Fields{
		"buildDate":          build.Date,
		"buildPlatform":      build.Platform,
		"buildSimpleVersion": build.SimpleVersion,
		"buildGitVersion":    build.GitVersion,
		"buildGitCommit":     build.GitCommit,
		"buildImageVersion":  build.ImageVersion,
	})

	var ok bool
	if nodeName, ok = os.LookupEnv("PG_NOMINEE_NODE_NAME"); !ok {
		logger.Fatalf("You must specify the env var PG_NOMINEE_NODE_NAME to a non-empty value.")
	}

	if nodeAddress, ok = os.LookupEnv("PG_NOMINEE_NODE_ADDRESS"); !ok {
		logger.Fatalf("You must specify the env var PG_NOMINEE_NODE_ADDRESS to a non-empty value (IP or DNS).")
	}

	if clusterName, ok = os.LookupEnv("PG_NOMINEE_CLUSTER_NAME"); !ok {
		logger.Fatalf("You must specify the env var PG_NOMINEE_CLUSTER_NAME to a non-empty value.")
	}

	if etcdEndPoints, ok = os.LookupEnv("PG_NOMINEE_ETCD_ENDPOINTS"); !ok {
		logger.Fatalf("You must specify the env var PG_NOMINEE_ETCD_ENDPOINTS to a non-empty value.")
	}

	if postgresPassword, ok = os.LookupEnv("POSTGRES_PASSWORD"); !ok {
		logger.Fatalf("You must specify the env var PG_NOMINEE_ETCD_ENDPOINTS to a non-empty value.")
	}
}

func main() {
	pg, _ := postgres.NewPostgres(
		nominee.Nominee{
			Name:    fmt.Sprintf("%s-%d", nodeName, time.Now().Nanosecond()), //Make sure the nodes do not collide. It is the pgnominee.main responsibility ?
			Cluster: clusterName,
			Address: nodeAddress,
		},
		postgres.DBUser{
			Username: "replicator",
			Password: "isgrfihgfiwhcfniw",
		},
		postgresPassword,
	)

	elector := etcdnode.NewEtcdNode(pg, strings.Split(etcdEndPoints, ","))
	defer elector.Cleanup()

	logger.Infof("starting...")
	if err := elector.Run(); err != nil {
		logger.Errorf("pgnominee: %v \n", err)
		return
	}

	<-elector.StopCh()
	logger.Infof("pgnominee: stopped.")
}
