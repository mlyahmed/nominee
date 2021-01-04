package election

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

type LeaderObserver interface {
	UpdateLeader(leader *node.Spec) error
}

// Cleaner ...
type Cleaner interface {
	Cleanup()
}

// Elector ...
type Elector interface {
	LeaderObserver
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
