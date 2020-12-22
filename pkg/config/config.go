package config

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Loader interface {
	LoadConfig(ctx context.Context)
}

type BasicConfig struct {
	Cluster string
	Domain  string
	loaded  bool
}

func init() {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(false)

	if os.Getenv("NOMINEE_ENVIRONMENT") == "DEV" {
		viper.SetConfigName("dev")
		viper.SetConfigType("env")
		viper.AddConfigPath(filepath.Dir("") + "/hack")
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}

}

func NewBasicConfig() *BasicConfig {
	return &BasicConfig{}
}

func (conf *BasicConfig) LoadConfig(context.Context) {
	if conf.loaded {
		return
	}
	conf.Cluster = GetStringOrPanic("NOMINEE_CLUSTER_NAME")
	conf.Domain = GetStringOrPanic("NOMINEE_DOMAIN_NAME")
	conf.loaded = true
}

func GetStringOrPanic(key string) string {
	if viper.GetString(key) == "" {
		panic(fmt.Sprintf("You must specify the env var %s to a non-empty value.", key))
	}
	return viper.GetString(key)
}

func GetIntOrPanic(key string) int {
	return viper.GetInt(key)
}

func SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

func GetString(key string) string {
	return viper.GetString(key)
}
