package main

import (
	"context"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/postgres"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	basicConfig := config.NewBasicConfig()
	pgConfig := postgres.NewPostgresConfig(basicConfig)
	pgConfig.LoadConfig(context.TODO())
	etcdConfig := etcd.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(context.TODO())
	node := postgres.NewPostgres(pgConfig)
	elector := etcd.NewElector(etcdConfig)
	_ = runner.RunElector(context.Background(), elector, node)
}
