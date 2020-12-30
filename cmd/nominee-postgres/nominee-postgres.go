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

	basicConfig := config.NewBasicConfig()

	pgConfig := service.NewPostgresConfig(basicConfig)
	pgConfig.LoadConfig(context.Background())

	etcdConfig := etcdconfig.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(context.Background())

	etcdRacer := etcd.NewEtcdRacer(etcdConfig)
	defer etcdRacer.Cleanup()

	log.Infof("starting...")
	if err := etcdRacer.Run(service.NewPostgres(pgConfig)); err != nil {
		log.Errorf("pgnominee: %v \n", err)
		return
	}

	<-etcdRacer.Stop()
	log.Infof("pgnominee: stopped.")
}
