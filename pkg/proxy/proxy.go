package proxy

import "github/mlyahmed.io/nominee/pkg/node"

// Proxy ...
type Proxy interface {
	PushNodes(nodes ...node.Spec) error
	PushLeader(leader node.Spec) error
	RemoveNode(electionKey string) error
}
