package runner

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

func RunElector(ctx context.Context, elector election.Elector, node nominee.Node) error {
	log := logger.G(ctx)
	defer elector.Cleanup()
	log.Infof("RunElector: starting...")
	if err := elector.Run(node); err != nil {
		log.Errorf("RunElector: %v \n", err)
		return err
	}
	<-elector.Stop()
	log.Infof("RunElector: stopped.")
	return nil
}

func RunObserver(ctx context.Context, observer election.Observer, proxy nominee.Proxy) error {
	log := logger.G(ctx)
	defer observer.Cleanup()
	log.Infof("RunObserver: starting...")
	if err := observer.Observe(proxy); err != nil {
		log.Errorf("RunObserver: %v \n", err)
		return err
	}
	<-observer.Stop()
	log.Infof("RunObserver: stopped.")
	return nil
}
