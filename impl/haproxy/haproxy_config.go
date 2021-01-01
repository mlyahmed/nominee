package haproxy

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
)

// HAProxyConfig ...
type HAProxyConfig struct {
	*config.BasicConfig
	ConfigFile string
	ExecFile   string
	TxDir      string
}

// NewHAProxyConfig ...
func NewHAProxyConfig(basic *config.BasicConfig) *HAProxyConfig {
	return &HAProxyConfig{BasicConfig: basic}
}

// LoadConfig ...
func (conf *HAProxyConfig) LoadConfig(ctx context.Context) {
	conf.BasicConfig.LoadConfig(ctx)

	config.SetDefault("NOMINEE_HAPROXY_CONFIG_FILE", "/usr/local/etc/haproxy/haproxy.cfg")
	config.SetDefault("NOMINEE_HAPROXY_EXEC_FILE", "/usr/local/sbin/haproxy")
	config.SetDefault("NOMINEE_HAPROXY_TX_DIR", "/tmp/haproxy")

	conf.ConfigFile = config.GetString("NOMINEE_HAPROXY_CONFIG_FILE")
	conf.ExecFile = config.GetString("NOMINEE_HAPROXY_EXEC_FILE")
	conf.TxDir = config.GetString("NOMINEE_HAPROXY_TX_DIR")
}
