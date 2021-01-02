package main

import (
	"context"
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/haproxy"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	haproxyConfig := haproxy.NewConfigLoader()
	haproxyConfig.Load(context.TODO())

	etcdConfig := etcd.NewConfigLoader()
	_ = runner.RunObserver(context.Background(), etcd.NewEtcdObserver(etcdConfig), haproxy.NewHAProxy(haproxyConfig))
}
