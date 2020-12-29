package main

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/race/etcd"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"github/mlyahmed.io/nominee/pkg/service"
)

func main() {
	log := logger.G(context.Background())
	etcdConfig := etcdconfig.NewEtcdConfig(config.NewBasicConfig())
	etcdConfig.LoadConfig(context.Background())
	etcdRacer := etcd.NewEtcdRacer(etcdConfig)
	defer etcdRacer.Cleanup()

	log.Infof("starting...")
	if err := etcdRacer.Run(service.NewDummy()); err != nil {
		log.Errorf("dummy: %v \n", err)
		return
	}

	<-etcdRacer.StopChan()
	log.Infof("dummy: stopped.")
}
