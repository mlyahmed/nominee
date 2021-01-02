package postgres

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/node"
	"os"
)

// ConfigLoader ...
type ConfigLoader interface {
	config.Loader
	GetSpec() *ConfigSpec
}

// ConfigSpec ...
type ConfigSpec struct {
	*config.BasicConfig
	NodeSpec   node.Spec
	Postgres   DBUser
	Replicator DBUser
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &ConfigSpec{
		BasicConfig: config.NewBasicConfig(),
		NodeSpec:    node.Spec{},
		Postgres:    DBUser{},
		Replicator:  DBUser{},
	}
}

// LoadConfig ...
func (conf *ConfigSpec) Load(ctx context.Context) {
	conf.BasicConfig.Load(ctx)
	config.SetDefault("NOMINEE_POSTGRES_NODE_PORT", 5432)

	conf.NodeSpec.Name = config.GetStringOrPanic("NOMINEE_POSTGRES_NODE_NAME")
	conf.NodeSpec.Address = config.GetStringOrPanic("NOMINEE_POSTGRES_NODE_ADDRESS")
	conf.NodeSpec.Port = int64(config.GetIntOrPanic("NOMINEE_POSTGRES_NODE_PORT"))
	conf.Postgres.Password = config.GetStringOrPanic("NOMINEE_POSTGRES_PASSWORD")
	conf.Replicator.Username = config.GetStringOrPanic("NOMINEE_POSTGRES_REP_USERNAME")
	conf.Replicator.Password = config.GetStringOrPanic("NOMINEE_POSTGRES_REP_PASSWORD")

	if err := os.Setenv("POSTGRES_PASSWORD", conf.Postgres.Password); err != nil {
		panic(err)
	}
}

func (conf *ConfigSpec) GetSpec() *ConfigSpec {
	return conf
}
