package election

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"reflect"
	"testing"
)

type electorSuite struct{}

var (
	nodeSpecExamples = []*node.Spec{
		{Name: "Node-001", Address: "192.168.1.1", Port: 2222},
		{Name: "Node-002", Address: "172.10.0.21", Port: 8989},
		{Name: "Node-003", Address: "10.10.0.1", Port: 5432},
	}
)

func TestElector(t *testing.T, factory func() Elector) {
	suite := electorSuite{}
	tests := []struct {
		description string
		run         func(*testing.T, func() Elector)
	}{
		{"when start then keep running", suite.whenRunWithoutErrorThenKeepRunning},
		{"when the node stops then stonith", suite.whenTheNodeStopsThenStonith},
		{"when elected then promote the node", suite.whenElectedThenPromoteTheNode},
		{"when error on promote then stonith", suite.whenErrorOnPromoteThenStonith},
		{"when demoted then stonith", suite.whenDemotedThenStonith},
		{"when another node is promoted then follow it", suite.whenAnotherNodeIsPromotedThenFollowIt},
		{"when error on follow then stonith", suite.whenErrorOnFollowThenStonith},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.run(t, factory)
		})
	}
}

func (electorSuite) whenRunWithoutErrorThenKeepRunning(t *testing.T, factory func() Elector) {
	elector := factory()
	defer elector.Cleanup()
	if err := elector.Run(mock.NewNode(t, &node.Spec{})); err != nil {
		t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
	}
	testutils.ItMustKeepRunning(t, elector.Done())
}

func (electorSuite) whenTheNodeStopsThenStonith(t *testing.T, factory func() Elector) {
	elector := factory()
	defer elector.Cleanup()
	nod := mock.NewNode(t, &node.Spec{})
	nod.StonithFn = func(context.Context) {
		close(nod.StopChan)
	}

	if err := elector.Run(nod); err != nil {
		t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
	}

	nod.Stonith(context.Background())

	testutils.ItMustBeStopped(t, elector.Done())
}

func (electorSuite) whenElectedThenPromoteTheNode(t *testing.T, factory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := factory()
			defer elector.Cleanup()

			nod := mock.NewNode(t, example)
			nod.LeadFn = func(_ context.Context, spec node.Spec) error {
				if reflect.DeepEqual(example, spec) {
					t.Fatalf("\t\t%spec FATAL: Elector, expected to <%v> as leader. Actual <%v>", testutils.Failed, example, spec)
				}
				return nil
			}

			if err := elector.Run(nod); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(example); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if nod.LeadHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to promote the node. Actally not.", testutils.Failed)
			}
		})
	}
}

func (electorSuite) whenErrorOnPromoteThenStonith(t *testing.T, factory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := factory()
			defer elector.Cleanup()

			nod := mock.NewNode(t, example)
			nod.LeadFn = func(_ context.Context, spec node.Spec) error {
				return errors.New("")
			}

			if err := elector.Run(nod); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(example); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			//FIXME: must stonith the node also

			testutils.ItMustBeStopped(t, elector.Done())
		})
	}
}

func (electorSuite) whenDemotedThenStonith(t *testing.T, factory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := factory()
			defer elector.Cleanup()
			nod := mock.NewNode(t, example)
			nod.LeadFn = func(_ context.Context, spec node.Spec) error { return nil }
			nod.StonithFn = func(context.Context) {}

			if err := elector.Run(nod); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(example); err != nil { // promote it
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&node.Spec{Name: string(uuid.NodeID())}); err != nil { // Demote it
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if nod.StonithHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to stonith the node.", testutils.Failed)
			}

			testutils.ItMustBeStopped(t, elector.Done())
		})
	}
}

func (electorSuite) whenAnotherNodeIsPromotedThenFollowIt(t *testing.T, factory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := factory()
			defer elector.Cleanup()
			nod := mock.NewNode(t, example)
			leader := node.Spec{Name: string(uuid.NodeID())}
			nod.FollowFn = func(ctx context.Context, l node.Spec) error {
				if leader != l {
					t.Fatalf("\t\t%s FATAL: Elector, expected <%v> as leader. Actual <%v>", testutils.Failed, leader, l)
				}
				return nil
			}

			if err := elector.Run(nod); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&leader); err != nil { // promote another node
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if nod.FollowHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to follow the leader. Actally not.", testutils.Failed)
			}
		})
	}
}

func (electorSuite) whenErrorOnFollowThenStonith(t *testing.T, factory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := factory()
			defer elector.Cleanup()
			nod := mock.NewNode(t, example)
			leader := node.Spec{Name: string(uuid.NodeID())}
			nod.FollowFn = func(ctx context.Context, l node.Spec) error {
				return errors.New("")
			}

			nod.StonithFn = func(context.Context) {}

			if err := elector.Run(nod); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&leader); err != nil { // Promote another node
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if nod.StonithHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to stonith the node.", testutils.Failed)
			}

			testutils.ItMustBeStopped(t, elector.Done())
		})
	}
}
