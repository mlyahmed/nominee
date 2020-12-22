package main

import (
	"context"
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/race"
)

func main() {
	log := logger.G(context.Background())
	ctx := context.Background()

	basicConfig := config.NewBasicConfig()

	haproxyConfig := proxy.NewHAProxyConfig(basicConfig)
	haproxyConfig.LoadConfig(ctx)

	etcdConfig := race.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(ctx)

	observer := race.NewEtcdObserver(etcdConfig)
	defer observer.Cleanup()

	log.Infof("starting...")
	if err := observer.Observe(proxy.NewHAProxy(haproxyConfig)); err != nil {
		log.Errorf("proxynominee: %v \n", err)
		return
	}

	<-observer.StopChan()
	log.Infof("proxynominee: stopped.")
}
