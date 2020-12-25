package race

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
)

// EtcdConfig ...
type EtcdConfig struct {
	*config.BasicConfig
	Endpoints string
	Username  string
	Password  string
}

// NewEtcdConfig ...
func NewEtcdConfig(basic *config.BasicConfig) *EtcdConfig {
	return &EtcdConfig{
		BasicConfig: basic,
	}
}

// LoadConfig ...
func (conf *EtcdConfig) LoadConfig(ctx context.Context) {
	conf.BasicConfig.LoadConfig(ctx)
	conf.Endpoints = config.GetStringOrPanic("NOMINEE_ETCD_ENDPOINTS")
	conf.Username = config.GetString("NOMINEE_ETCD_USERNAME")
	conf.Password = config.GetString("NOMINEE_ETCD_PASSWORD")
}
