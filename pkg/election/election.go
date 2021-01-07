package election

import (
	"github/mlyahmed.io/nominee/pkg/node"
)

type LeaderWatcher interface {
	UpdateLeader(leader *node.Spec) error
}

type NodesWatcher interface {
	UpdateNodes(nodes []*node.Spec) error
	RemoveNodes(nodes ...*node.Spec) error
}
