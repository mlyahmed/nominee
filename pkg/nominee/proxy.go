package nominee

// Proxy ...
type Proxy interface {
	PushNodes(nodes ...NodeSpec) error
	PushLeader(leader NodeSpec) error
	RemoveNode(electionKey string) error
}
