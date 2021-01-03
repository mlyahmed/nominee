package main

import (
	"context"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/postgres"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	elector := etcd.NewElector(etcd.NewConfigLoader())
	node := postgres.NewPostgres(postgres.NewConfigLoader())
	rn := runner.NewElectorRunner()
	if err := rn.Run(context.Background(), elector, node); err != nil {
		logger.G(context.TODO()).Fatalf("PostgresW: Failed to run: %v", err)
	}
}
