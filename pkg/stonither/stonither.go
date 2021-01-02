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
	Ctx       context.Context
	CancelFn  func()
	StopChan  chan struct{}
	ErrorChan chan error
	Status
}

func NewBase() *Base {
	ctx, cancel := context.WithCancel(context.Background())
	base := &Base{
		Ctx:       ctx,
		CancelFn:  cancel,
		StopChan:  make(chan struct{}),
		ErrorChan: make(chan error),
	}
	base.setUpSignals()
	base.setUpChannels()
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

func (base *Base) setUpChannels() {
	go func() {
		for err := range base.ErrorChan {
			if err != nil {
				base.Stonith()
			}
		}
	}()
}

func (base *Base) Stonith() {
	base.CancelFn()
	select {
	case _, ok := <-base.StopChan:
		if ok {
			close(base.StopChan)
		}
	default:
		close(base.StopChan)
	}
}

func (base *Base) Reset() {
	base.CancelFn()
	base.Ctx, base.CancelFn = context.WithCancel(context.Background())
}

func (base *Base) Done() node.StopChan {
	return base.StopChan
}
