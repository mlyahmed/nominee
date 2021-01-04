package election

import (
	"context"
	"github.com/google/uuid"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"testing"
)

type electorSuite struct{}

func TestElector(t *testing.T, electorFactory func() Elector) {
	suite := electorSuite{}
	tests := []struct {
		description string
		run         func(*testing.T, func() Elector)
	}{
		{"when start then keep running", suite.whenRunWithoutErrorThenKeepRunning},
		{"when the node stops then stonith", suite.whenTheNodeStopsThenStonith},
		{"when elected then promote the node", suite.whenElectedThenPromoteTheNode},
		{"when demoted then stonith", suite.whenDemotedThenStonith},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.run(t, electorFactory)
		})
	}
}

func (electorSuite) whenRunWithoutErrorThenKeepRunning(t *testing.T, electorFactory func() Elector) {
	e := electorFactory()
	defer e.Cleanup()
	if err := e.Run(mock.NewNode(t, &node.Spec{})); err != nil {
		t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
	}
	testutils.ItMustKeepRunning(t, e.Done())
}

func (electorSuite) whenTheNodeStopsThenStonith(t *testing.T, electorFactory func() Elector) {
	e := electorFactory()
	defer e.Cleanup()
	n := mock.NewNode(t, &node.Spec{})
	n.StonithFn = func(context.Context) error {
		close(n.StopChan)
		return nil
	}

	if err := e.Run(n); err != nil {
		t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
	}

	if err := n.Stonith(context.Background()); err != nil {
		t.Fatalf("\t\t%s FATAL: Elector, error when stonith the node %v", testutils.Failed, err)
	}

	testutils.ItMustBeStopped(t, e.Done())

}

var (
	nodeSpecExamples = []node.Spec{
		{Name: "Node-001", Address: "192.168.1.1", Port: 2222},
		{Name: "Node-002", Address: "172.10.0.21", Port: 8989},
		{Name: "Node-003", Address: "10.10.0.1", Port: 5432},
	}
)

func (electorSuite) whenElectedThenPromoteTheNode(t *testing.T, electorFactory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := electorFactory()
			defer elector.Cleanup()

			n := mock.NewNode(t, &example)
			n.LeadFn = func(_ context.Context, s node.Spec) error {
				if example != s {
					t.Fatalf("\t\t%s FATAL: Elector, expected to <%v> as leader. Actual <%v>", testutils.Failed, example, s)
				}
				return nil
			}

			if err := elector.Run(n); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&example); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if n.LeadHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to promote the node. Actally not.", testutils.Failed)
			}
		})
	}
}

func (electorSuite) whenDemotedThenStonith(t *testing.T, electorFactory func() Elector) {
	for _, example := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
			elector := electorFactory()
			defer elector.Cleanup()
			n := mock.NewNode(t, &example)
			n.LeadFn = func(_ context.Context, s node.Spec) error { return nil }
			n.StonithFn = func(context.Context) error { return nil }

			if err := elector.Run(n); err != nil {
				t.Fatalf("\t\t%s FATAL: Elector, failed to run %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&example); err != nil { // promote it
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if err := elector.UpdateLeader(&node.Spec{Name: string(uuid.NodeID())}); err != nil { // Demote it
				t.Fatalf("\t\t%s FATAL: Elector, failed to update leader %v", testutils.Failed, err)
			}

			if n.StonithHits != 1 {
				t.Fatalf("\t\t%s FATAL: Elector, expected to stonith the node.", testutils.Failed)
			}

			testutils.ItMustBeStopped(t, elector.Done())
		})
	}
}
