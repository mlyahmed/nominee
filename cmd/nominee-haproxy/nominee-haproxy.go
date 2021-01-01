package main

import (
	"context"
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/haproxy"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	basicConfig := config.NewBasicConfig()
	haproxyConfig := haproxy.NewHAProxyConfig(basicConfig)
	haproxyConfig.LoadConfig(context.TODO())
	etcdConfig := etcd.NewEtcdConfig(basicConfig)
	etcdConfig.LoadConfig(context.TODO())
	observer := etcd.NewEtcdObserver(etcdConfig)
	proxy := haproxy.NewHAProxy(haproxyConfig)
	_ = runner.RunObserver(context.Background(), observer, proxy)
}
