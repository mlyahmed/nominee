package haproxy

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
)

// ConfigLoader ...
type ConfigLoader interface {
	config.Loader
	GetSpec() *ConfigSpec
}

// ConfigSpec ...
type ConfigSpec struct {
	*config.BasicConfig
	ConfigFile string
	ExecFile   string
	TxDir      string
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &ConfigSpec{BasicConfig: config.NewBasicConfig()}
}

// LoadConfig ...
func (conf *ConfigSpec) Load(ctx context.Context) {
	conf.BasicConfig.Load(ctx)

	config.SetDefault("NOMINEE_HAPROXY_CONFIG_FILE", "/usr/local/etc/haproxy/haproxy.cfg")
	config.SetDefault("NOMINEE_HAPROXY_EXEC_FILE", "/usr/local/sbin/haproxy")
	config.SetDefault("NOMINEE_HAPROXY_TX_DIR", "/tmp/haproxy")

	conf.ConfigFile = config.GetString("NOMINEE_HAPROXY_CONFIG_FILE")
	conf.ExecFile = config.GetString("NOMINEE_HAPROXY_EXEC_FILE")
	conf.TxDir = config.GetString("NOMINEE_HAPROXY_TX_DIR")
}

func (conf *ConfigSpec) GetSpec() *ConfigSpec {
	return conf
}
