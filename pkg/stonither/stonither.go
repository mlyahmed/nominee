package stonither

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/node"
	"os"
	"os/signal"
)

type Status int

type Stonither interface {
	Stonith()
	Reset()
	Done() node.StopChan
}

type Base struct {
	Ctx      context.Context
	CancelFn func()
	Status
	stopChan chan struct{}
}

func NewBase() *Base {
	ctx, cancel := context.WithCancel(context.Background())
	base := &Base{
		Ctx:      ctx,
		CancelFn: cancel,
		stopChan: make(chan struct{}),
	}
	base.setUpSignals()
	return base
}

func (base *Base) setUpSignals() {
	listener := make(chan os.Signal, len(ShutdownSignals))
	signal.Notify(listener, ShutdownSignals...)
	go func() {
		<-listener
		base.Stonith()
		<-listener
		os.Exit(1)
	}()
}

func (base *Base) Stonith() {
	base.CancelFn()
	select {
	case _, ok := <-base.stopChan:
		if ok {
			close(base.stopChan)
		}
	default:
		close(base.stopChan)
	}
}

func (base *Base) Reset() {
	base.CancelFn()
	base.Ctx, base.CancelFn = context.WithCancel(context.Background())
}

func (base *Base) Done() node.StopChan {
	return base.stopChan
}
