package main

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/race"
	"github/mlyahmed.io/nominee/pkg/service"
)

func main() {
	log := logger.G(context.Background())

	basicConfig := config.NewBasicConfig()

	pgConfig := service.NewPostgresConfig(basicConfig)
	pgConfig.LoadConfig(context.Background())

	etcdConfig := race.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(context.Background())

	etcd := race.NewEtcdRacer(etcdConfig)
	defer etcd.Cleanup()

	log.Infof("starting...")
	if err := etcd.Run(service.NewPostgres(pgConfig)); err != nil {
		log.Errorf("pgnominee: %v \n", err)
		return
	}

	<-etcd.StopChan()
	log.Infof("pgnominee: stopped.")
}
