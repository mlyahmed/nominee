package mock

import (
	"context"
	"github/mlyahmed.io/nominee/infra"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"testing"
)

// NodeRecord ...
type NodeRecord struct {
	LeadHits    int
	FollowHits  int
	StonithHits int
	Leader      nominee.NodeSpec
}

// Node ...
type Node struct {
	*nominee.NodeBase
	*NodeRecord
	t            *testing.T
	StopChan     chan struct{}
	DaemonNameFn func() string
	LeadFn       func(context.Context, nominee.NodeSpec) error
	FollowFn     func(context.Context, nominee.NodeSpec) error
	StonithFn    func(context.Context) error
	StopChanFn   func() nominee.StopChan
}

// NewMockServiceWithNominee ...
func NewNode(t *testing.T, node *nominee.NodeSpec) *Node {
	stopChan := make(chan struct{}, 1)
	return &Node{
		NodeBase: &nominee.NodeBase{
			NodeSpec: node,
		},
		NodeRecord: &NodeRecord{},
		StopChan:   stopChan,
		DaemonNameFn: func() string {
			return "mockedService"
		},
		LeadFn: func(ctx context.Context, nominee nominee.NodeSpec) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: LeadFn function not specified.", infra.Failed)
			return nil
		},
		FollowFn: func(ctx context.Context, n nominee.NodeSpec) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: FollowFn function not specified.", infra.Failed)
			return nil
		},
		StonithFn: func(ctx context.Context) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: StonithFn function not specified.", infra.Failed)
			return nil
		},
		StopChanFn: func() nominee.StopChan {
			return stopChan
		},
	}
}

// DaemonName ...
func (mock *Node) DaemonName() string {
	return mock.DaemonNameFn()
}

// Lead ...
func (mock *Node) Lead(ctx context.Context, leader nominee.NodeSpec) error {
	mock.LeadHits++
	mock.Leader = leader
	return mock.LeadFn(ctx, leader)
}

// Follow ...
func (mock *Node) Follow(ctx context.Context, leader nominee.NodeSpec) error {
	mock.FollowHits++
	mock.Leader = leader
	return mock.FollowFn(ctx, leader)
}

// Stonith ...
func (mock *Node) Stonith(ctx context.Context) error {
	mock.StonithHits++
	return mock.StonithFn(ctx)
}

// StopChan ...
func (mock *Node) Stop() nominee.StopChan {
	return mock.StopChanFn()
}
