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
	RunFn func(p proxy.Proxy) error
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

func (c *Cleaner) Cleanup() {
	c.CleanupFn()
}

func (e *Elector) Run(n node.Node) error {
	return e.RunFn(n)
}

func (o *Observer) Run(p proxy.Proxy) error {
	return o.RunFn(p)
}
