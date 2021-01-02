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
	stonither.Stonither
	Cleaner
	Run(node.Node) error
}

// Observer ...
type Observer interface {
	stonither.Stonither
	Cleaner
	Observe(proxy proxy.Proxy) error
}
