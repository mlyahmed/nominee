package config

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
)

// Loader ...
type Loader interface {
	Load(ctx context.Context)
}

// BasicConfig ...
type BasicConfig struct {
	Cluster string
	Domain  string
	loaded  bool
}

var setup bool

func setItUp() {
	if setup {
		return
	}
	viper.AutomaticEnv()
	if viper.IsSet("NOMINEE_CONF_FILE") {
		file := viper.GetString("NOMINEE_CONF_FILE")
		parts := strings.Split(path.Base(file), ".")
		viper.SetConfigName(parts[0])
		viper.SetConfigType(parts[1])
		viper.AddConfigPath(path.Dir(file))
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}
	setup = true
}

// NewBasicConfig ...
func NewBasicConfig() *BasicConfig {
	return &BasicConfig{}
}

// LoadConfig ...
func (conf *BasicConfig) Load(context.Context) {
	if conf.loaded {
		return
	}
	conf.Cluster = GetStringOrPanic("NOMINEE_CLUSTER_NAME")
	conf.Domain = GetStringOrPanic("NOMINEE_DOMAIN_NAME")
	conf.loaded = true
}

// GetStringOrPanic ...
func GetStringOrPanic(key string) string {
	setItUp()
	if viper.GetString(key) == "" {
		panic(fmt.Sprintf("You must specify the env var %s to a non-empty value.", key))
	}
	return viper.GetString(key)
}

// GetIntOrPanic ...
func GetIntOrPanic(key string) int {
	setItUp()
	if viper.GetString(key) == "" {
		panic(fmt.Sprintf("You must specify the env var %s to a valid int value.", key))
	}
	return viper.GetInt(key)
}

// SetDefault ...
func SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

// GetString ...
func GetString(key string) string {
	setItUp()
	return viper.GetString(key)
}

// Reset ...
func Reset() {
	for _, key := range append(viper.AllKeys(), "NOMINEE_CONF_FILE") {
		_ = os.Unsetenv(key)
	}
	viper.Reset()
	setup = false
}
