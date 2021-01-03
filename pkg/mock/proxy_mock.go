package mock

import "github/mlyahmed.io/nominee/pkg/node"

type Proxy struct {
	PushNodesFn  func(nodes ...node.Spec) error
	PushLeaderFn func(leader node.Spec) error
	RemoveNodeFn func(electionKey string) error
}

func NewProxy() *Proxy {
	return &Proxy{
		PushNodesFn: func(nodes ...node.Spec) error {
			return nil
		},
		PushLeaderFn: func(leader node.Spec) error {
			return nil
		},
		RemoveNodeFn: func(electionKey string) error {
			return nil
		},
	}
}

func (p *Proxy) PushNodes(nodes ...node.Spec) error {
	return p.PushNodesFn(nodes...)
}

func (p *Proxy) PushLeader(leader node.Spec) error {
	return p.PushLeaderFn(leader)
}

func (p *Proxy) RemoveNode(electionKey string) error {
	return p.RemoveNodeFn(electionKey)
}
