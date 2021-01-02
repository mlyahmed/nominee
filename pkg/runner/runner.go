package runner

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
)

func RunElector(ctx context.Context, elector election.Elector, node node.Node) error {
	log := logger.G(ctx)
	defer elector.Cleanup()
	log.Infof("RunElector: starting...")
	if err := elector.Run(node); err != nil {
		log.Errorf("RunElector: %v", err)
		return err
	}
	<-elector.Done()
	log.Infof("RunElector: stopped.")
	return nil
}

func RunObserver(ctx context.Context, observer election.Observer, proxy proxy.Proxy) error {
	log := logger.G(ctx)
	defer observer.Cleanup()
	log.Infof("RunObserver: starting...")
	if err := observer.Observe(proxy); err != nil {
		log.Errorf("RunObserver: %v", err)
		return err
	}
	<-observer.Done()
	log.Infof("RunObserver: stopped.")
	return nil
}
