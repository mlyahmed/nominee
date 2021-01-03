package election

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

// Cleaner ...
type Cleaner interface {
	Cleanup()
}

// Elector ...
type Elector interface {
	Run(node.Node) error
	stonither.Stonither
	Cleaner
}

// Observer ...
type Observer interface {
	Observe(proxy.Proxy) error
	stonither.Stonither
	Cleaner
}
