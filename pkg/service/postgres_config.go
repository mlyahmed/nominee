package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

type PGConfig struct {
	*config.BasicConfig
	Nominee    nominee.Nominee
	Postgres   DBUser
	Replicator DBUser
}

func NewPostgresConfig(basic *config.BasicConfig) *PGConfig {
	return &PGConfig{
		BasicConfig: basic,
		Nominee:     nominee.Nominee{},
		Postgres:    DBUser{},
		Replicator:  DBUser{},
	}
}

func (conf *PGConfig) LoadConfig(ctx context.Context) {
	conf.BasicConfig.LoadConfig(ctx)
	config.SetDefault("NOMINEE_POSTGRES_NODE_PORT", 5432)

	conf.Nominee.Name = config.GetStringOrPanic("NOMINEE_POSTGRES_NODE_NAME")
	conf.Nominee.Address = config.GetStringOrPanic("NOMINEE_POSTGRES_NODE_ADDRESS")
	conf.Nominee.Port = int64(config.GetIntOrPanic("NOMINEE_POSTGRES_NODE_PORT"))
	conf.Postgres.Password = config.GetStringOrPanic("NOMINEE_POSTGRES_POSTGRES_PASSWORD")
	conf.Replicator.Username = config.GetStringOrPanic("NOMINEE_POSTGRES_REP_USERNAME")
	conf.Replicator.Password = config.GetStringOrPanic("NOMINEE_POSTGRES_REP_PASSWORD")
}
