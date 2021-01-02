package main

import (
	"context"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/postgres"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	pgConfig := postgres.NewConfigLoader()
	etcdConfig := etcd.NewConfigLoader()
	_ = runner.RunElector(context.Background(), etcd.NewElector(etcdConfig), postgres.NewPostgres(pgConfig))
}
