package main

import (
	"context"
	_ "github.com/haproxytech/client-native/v2/runtime"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/impl/haproxy"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/runner"
)

func main() {
	observer := etcd.NewEtcdObserver(etcd.NewConfigLoader())
	proxy := haproxy.NewHAProxy(haproxy.NewConfigLoader())
	or := runner.NewObserverRunner()
	if err := or.Run(context.Background(), observer, proxy); err != nil {
		logger.G(context.TODO()).Fatalf("HAProxyW: Failed to run: %v", err)
	}
}
