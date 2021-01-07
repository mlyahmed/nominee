package stonither

import (
	"context"
	basepkg "github/mlyahmed.io/nominee/pkg/base"
	"os"
	"os/signal"
)

type Status int

type Stonither interface {
	basepkg.Doer
	Stonith(context.Context)
}

type Base struct {
	Ctx      context.Context
	CancelFn func()
	Status
	doneChan chan struct{}
}

func NewBase() *Base {
	ctx, cancel := context.WithCancel(context.Background())
	base := &Base{
		Ctx:      ctx,
		CancelFn: cancel,
		doneChan: make(chan struct{}),
	}
	base.setUpSignals()
	return base
}

func (base *Base) setUpSignals() {
	listener := make(chan os.Signal, len(ShutdownSignals))
	signal.Notify(listener, ShutdownSignals...)
	go func() {
		<-listener
		base.Stonith(context.TODO())
		<-listener
		os.Exit(1)
	}()
}

func (base *Base) Stonith(context.Context) {
	base.CancelFn()
	select {
	case _, ok := <-base.doneChan:
		if ok {
			close(base.doneChan)
		}
	default:
		close(base.doneChan)
	}
}

// Reset see types.Reseter
func (base *Base) Reset() {
	base.CancelFn()
	base.Ctx, base.CancelFn = context.WithCancel(context.Background())
}

func (base *Base) Done() basepkg.DoneChan {
	return base.doneChan
}
