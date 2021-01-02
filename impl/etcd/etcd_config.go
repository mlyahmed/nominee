package etcd

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"strings"
)

type ConfigLoader interface {
	config.Loader
	GetSpec() *ConfigSpec
}

// ConfigSpec ...
type ConfigSpec struct {
	*config.BasicConfig
	Endpoints []string
	Username  string
	Password  string
	Loaded    bool
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &ConfigSpec{BasicConfig: config.NewBasicConfig()}
}

// LoadConfig ...
func (conf *ConfigSpec) Load(ctx context.Context) {
	conf.BasicConfig.Load(ctx)
	conf.Endpoints = strings.Split(config.GetStringOrPanic("NOMINEE_ETCD_ENDPOINTS"), ",")
	conf.Username = config.GetString("NOMINEE_ETCD_USERNAME")
	conf.Password = config.GetString("NOMINEE_ETCD_PASSWORD")
	conf.Loaded = true
}

func (conf *ConfigSpec) GetSpec() *ConfigSpec {
	if !conf.Loaded {
		panic("config not loaded.")
	}
	return conf
}
