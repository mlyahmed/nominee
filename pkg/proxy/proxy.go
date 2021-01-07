package proxy

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

// Proxy ...
type Proxy interface {
	Publish(leader *node.Spec, followers ...*node.Spec) error
	stonither.Stonither
}
