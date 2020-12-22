package race

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
)

type EtcdConfig struct {
	*config.BasicConfig
	endpoints string
	username  string
	password  string
}

func NewEtcdConfig(basic *config.BasicConfig) *EtcdConfig {
	return &EtcdConfig{
		BasicConfig: basic,
	}
}

func (conf *EtcdConfig) LoadConfig(ctx context.Context) {
	conf.BasicConfig.LoadConfig(ctx)
	conf.endpoints = config.GetStringOrPanic("NOMINEE_ETCD_ENDPOINTS")
	conf.username = config.GetString("NOMINEE_ETCD_USERNAME")
	conf.password = config.GetString("NOMINEE_ETCD_PASSWORD")
}
