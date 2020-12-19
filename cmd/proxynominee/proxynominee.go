package main

import (
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/build"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/race"
	"os"
	"strings"
)

var (
	logger    *logrus.Entry
	domain    string
	cluster   string
	endpoints string
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
	if cluster, ok = os.LookupEnv("NOMINEE_CLUSTER_NAME"); !ok {
		logger.Fatalf("You must specify the env var NOMINEE_CLUSTER_NAME to a non-empty value.")
	}

	if domain, ok = os.LookupEnv("NOMINEE_DOMAIN"); !ok {
		logger.Fatalf("You must specify the env var NOMINEE_DOMAIN to a non-empty value.")
	}

	if endpoints, ok = os.LookupEnv("NOMINEE_ETCD_ENDPOINTS"); !ok {
		logger.Fatalf("You must specify the env var NOMINEE_ETCD_ENDPOINTS to a non-empty value.")
	}

}

func main() {
	observer := race.NewEtcdObserver(strings.Split(endpoints, ","))
	defer observer.Cleanup()

	logger.Infof("starting...")
	if err := observer.Observe(proxy.NewHAProxy(domain, cluster)); err != nil {
		logger.Errorf("proxynominee: %v \n", err)
		return
	}

	<-observer.StopChan()
	logger.Infof("proxynominee: stopped.")
}
