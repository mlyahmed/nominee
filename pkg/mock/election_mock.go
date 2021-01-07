package mock

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/stonither"
	"testing"
)

type Mock struct {
	t *testing.T
}

type Cleaner struct {
	CleanupFn func()
}

type Elector struct {
	*Mock
	*stonither.Base
	*Cleaner
	RunFn func(node.Node) error
}

type Observer struct {
	*Mock
	*stonither.Base
	*Cleaner
	ObserveFn func(p proxy.Proxy) error
}

func NewElector(t *testing.T) *Elector {
	return &Elector{
		Mock: &Mock{t: t},
		Base: stonither.NewBase(),
		Cleaner: &Cleaner{
			CleanupFn: func() {},
		},
		RunFn: func(node.Node) error {
			return nil
		},
	}
}

func NewObserver(t *testing.T) *Observer {
	return &Observer{
		Mock: &Mock{t: t},
		Base: stonither.NewBase(),
		Cleaner: &Cleaner{
			CleanupFn: func() {},
		},
		ObserveFn: func(p proxy.Proxy) error {
			return nil
		},
	}
}

func (c *Cleaner) Cleanup() {
	c.CleanupFn()
}

func (e *Elector) Run(n node.Node) error {
	return e.RunFn(n)
}

func (e *Elector) UpdateLeader(_ *node.Spec) error {
	return nil
}

func (o *Observer) Observe(p proxy.Proxy) error {
	return o.ObserveFn(p)
}

func (o *Observer) UpdateLeader(*node.Spec) error {
	return nil
}

func (o *Observer) UpdateNodes([]*node.Spec) error {
	return nil
}

func (o *Observer) RemoveNodes(...*node.Spec) error {
	return nil
}
