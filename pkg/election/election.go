package election

import (
	"github/mlyahmed.io/nominee/pkg/nominee"
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
	Run(nominee.Node) error
}

// Observer ...
type Observer interface {
	stonither.Stonither
	Cleaner
	Observe(proxy nominee.Proxy) error
}
