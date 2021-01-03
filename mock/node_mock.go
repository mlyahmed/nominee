package mock

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"testing"
)

// NodeRecord ...
type NodeRecord struct {
	LeadHits    int
	FollowHits  int
	StonithHits int
	Leader      node.Spec
}

// Node ...
type Node struct {
	*node.Spec
	*NodeRecord
	t            *testing.T
	StopChan     chan struct{}
	DaemonNameFn func() string
	LeadFn       func(context.Context, node.Spec) error
	FollowFn     func(context.Context, node.Spec) error
	StonithFn    func(context.Context) error
	StopChanFn   func() node.StopChan
}

// NewMockServiceWithNominee ...
func NewNode(t *testing.T, spec *node.Spec) *Node {
	stopChan := make(chan struct{}, 1)
	return &Node{
		Spec:       spec,
		NodeRecord: &NodeRecord{},
		StopChan:   stopChan,
		DaemonNameFn: func() string {
			return "mockedService"
		},
		LeadFn: func(_ context.Context, _ node.Spec) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: LeadFn function not specified.", testutils.Failed)
			return nil
		},
		FollowFn: func(_ context.Context, _ node.Spec) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: FollowFn function not specified.", testutils.Failed)
			return nil
		},
		StonithFn: func(_ context.Context) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: StonithFn function not specified.", testutils.Failed)
			return nil
		},
		StopChanFn: func() node.StopChan {
			return stopChan
		},
	}
}

// DaemonName ...
func (mock *Node) GetDaemonName() string {
	return mock.DaemonNameFn()
}

// Lead ...
func (mock *Node) Lead(ctx context.Context, leader node.Spec) error {
	mock.LeadHits++
	mock.Leader = leader
	return mock.LeadFn(ctx, leader)
}

// Follow ...
func (mock *Node) Follow(ctx context.Context, leader node.Spec) error {
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
func (mock *Node) Stop() node.StopChan {
	return mock.StopChanFn()
}
