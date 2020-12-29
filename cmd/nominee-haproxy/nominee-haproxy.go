package main

import (
	"context"
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/race/etcd"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
)

func main() {
	log := logger.G(context.Background())
	ctx := context.Background()

	basicConfig := config.NewBasicConfig()

	haproxyConfig := proxy.NewHAProxyConfig(basicConfig)
	haproxyConfig.LoadConfig(ctx)

	etcdConfig := etcdconfig.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(ctx)

	etcdObserver := etcd.NewEtcdObserver(etcdConfig)
	defer etcdObserver.Cleanup()

	log.Infof("starting...")
	if err := etcdObserver.Observe(proxy.NewHAProxy(haproxyConfig)); err != nil {
		log.Errorf("proxynominee: %v \n", err)
		return
	}

	<-etcdObserver.StopChan()
	log.Infof("proxynominee: stopped.")
}
