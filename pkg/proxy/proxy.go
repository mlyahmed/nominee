package proxy

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

type Status int

const (
	Started Status = iota
	Stopped
)

// Proxy ...
type Proxy interface {
	Publish(leader *node.Spec, followers ...*node.Spec) error
	stonither.Stonither
}

type BasicProxy struct {
	*stonither.Basic
}

func NewBasicProxy() *BasicProxy {
	return &BasicProxy{
		Basic: stonither.NewBasic(),
	}
}
