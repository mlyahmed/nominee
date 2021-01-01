package stonither

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
	"os/signal"
)

type Status int

const (
	Started Status = iota
	Stopped
)

type Stonither interface {
	Stonith()
	Reset()
	Stop() nominee.StopChan
}

type Base struct {
	Context   context.Context
	CancelFn  func()
	StopChan  chan struct{}
	ErrorChan chan error
	Status
}

func NewBase() *Base {
	ctx, cancel := context.WithCancel(context.Background())
	base := &Base{
		Context:   ctx,
		CancelFn:  cancel,
		StopChan:  make(chan struct{}),
		ErrorChan: make(chan error),
	}
	base.setUpSignals()
	base.setUpChannels()
	base.Status = Started
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
	base.Status = Stopped
}

func (base *Base) Reset() {
	base.CancelFn()
	base.Context, base.CancelFn = context.WithCancel(context.Background())
}

func (base *Base) Stop() nominee.StopChan {
	return base.StopChan
}
