package etcdconfig

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"strings"
)

// Config ...
type Config struct {
	*config.BasicConfig
	Endpoints []string
	Username  string
	Password  string
}

// NewEtcdConfig ...
func NewEtcdConfig(basic *config.BasicConfig) *Config {
	return &Config{
		BasicConfig: basic,
	}
}

// LoadConfig ...
func (conf *Config) LoadConfig(ctx context.Context) {
	conf.BasicConfig.LoadConfig(ctx)
	conf.Endpoints = strings.Split(config.GetStringOrPanic("NOMINEE_ETCD_ENDPOINTS"), ",")
	conf.Username = config.GetString("NOMINEE_ETCD_USERNAME")
	conf.Password = config.GetString("NOMINEE_ETCD_PASSWORD")
}
