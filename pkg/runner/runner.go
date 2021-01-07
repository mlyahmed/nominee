package runner

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
)

type ElectorRunner interface {
	Run(context.Context, election.Elector, node.Node) error
}

type ObserverRunner interface {
	Run(ctx context.Context, observer election.Observer, proxy proxy.Proxy) error
}

type defaultElectorRunner struct{}
type defaultObserverRunner struct{}

func NewElectorRunner() ElectorRunner {
	return defaultElectorRunner{}
}

func NewObserverRunner() ObserverRunner {
	return defaultObserverRunner{}
}

func (defaultElectorRunner) Run(ctx context.Context, elector election.Elector, node node.Node) error {
	log := logger.G(ctx)
	defer elector.Cleanup()

	log.Infof("RunElector: starting...")
	if err := elector.Run(node); err != nil {
		log.Errorf("RunElector: %v", err)
		return err
	}

	select {
	case <-elector.Done():
		log.Infof("RunElector: done.")
	case <-ctx.Done():
		log.Infof("RunElector: context done.")
	}

	return nil
}

func (defaultObserverRunner) Run(ctx context.Context, observer election.Observer, proxy proxy.Proxy) error {
	log := logger.G(ctx)
	defer observer.Cleanup()
	log.Infof("RunObserver: starting...")
	if err := observer.Observe(proxy); err != nil {
		log.Errorf("RunObserver: %v", err)
		return err
	}

	select {
	case <-observer.Done():
		log.Infof("RunObserver: done.")
	case <-ctx.Done():
		log.Infof("RunObserver: context done.")
	}

	return nil
}
