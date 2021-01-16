package election

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"reflect"
	"sort"
	"testing"
)

type observerSuite struct{}

func TestObserver(t *testing.T, factory func() Observer) {
	suite := observerSuite{}
	tests := []struct {
		description string
		run         func(*testing.T, func() Observer)
	}{
		{"when start then keep running", suite.whenRunWithoutErrorThenKeepRunning},
		{"when the proxy stops then stonith", suite.whenTheProxyStopsThenStonith},
		{"when new leader then publish it", suite.whenNewLeaderThenPublishIt},
		{"when new node then add it as follower", suite.whenNewNodeThenAddItAsFollower},
		{"when a follower is removed then removes it", suite.whenAFollowerIsRemovedThenRemovesIt},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.run(t, factory)
		})
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.run(t, factory)
		})
	}
}

func (observerSuite) whenRunWithoutErrorThenKeepRunning(t *testing.T, factory func() Observer) {
	e := factory()
	defer e.Cleanup()

	if err := e.Observe(mock.NewProxy()); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	testutils.AsyncAssertion.ItMustKeepRunning(t, e.Done())
}

func (observerSuite) whenTheProxyStopsThenStonith(t *testing.T, factory func() Observer) {
	observer := factory()
	defer observer.Cleanup()
	proxy := mock.NewProxy()

	if err := observer.Observe(proxy); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	proxy.Stonith(context.TODO())

	testutils.AsyncAssertion.ItMustBeStopped(t, observer.Done())
}

func (observerSuite) whenNewLeaderThenPublishIt(t *testing.T, factory func() Observer) {
	for _, leader := range nodeSpecExamples {
		t.Run(leader.Name, func(t *testing.T) {
			observer := factory()
			defer observer.Cleanup()
			proxy := mock.NewProxy()

			if err := observer.Observe(proxy); err != nil {
				t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
			}

			if err := observer.UpdateLeader(leader); err != nil {
				t.Fatalf("\t\t%s FATAL: Observer, failed to update leader %v", testutils.Failed, err)
			}

			same := func() bool {
				return leader == proxy.Leader
			}
			testutils.AsyncAssertion.ItMustBeTrue(t, same)
		})
	}

}

func (observerSuite) whenNewFollowersThenPublishThem(t *testing.T, factory func() Observer) {
	observer := factory()
	defer observer.Cleanup()
	proxy := mock.NewProxy()
	if err := observer.Observe(proxy); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	if err := observer.UpdateNodes(nodeSpecExamples...); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to update nodes %v", testutils.Failed, err)
	}

	equal := func() bool {
		actual := make([]*node.Spec, len(proxy.Followers))

		for _, v := range proxy.Followers {
			actual = append(actual, v)
		}
		return reflect.DeepEqual(nodeSpecExamples, actual)
	}

	testutils.AsyncAssertion.ItMustBeTrue(t, equal)
}

func (observerSuite) whenNewNodeThenAddItAsFollower(t *testing.T, factory func() Observer) {
	observer := factory()
	defer observer.Cleanup()
	proxy := mock.NewProxy()
	if err := observer.Observe(proxy); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	if err := observer.UpdateNodes(nodeSpecExamples...); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to update nodes %v", testutils.Failed, err)
	}

	newNode := &node.Spec{ElectionKey: "new-key-001", Name: "new-node-001", Address: "27.23.56.98", Port: 4589}
	if err := observer.UpdateNodes(newNode); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to add the new node %v", testutils.Failed, err)
	}

	expected := append(nodeSpecExamples, newNode)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

	equal := func() bool {
		actual := make([]*node.Spec, len(proxy.Followers))
		i := 0
		for _, v := range proxy.Followers {
			actual[i] = v
			i++
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i].Name < actual[j].Name
		})

		return reflect.DeepEqual(expected, actual)
	}
	testutils.AsyncAssertion.ItMustBeTrue(t, equal)
}

func (observerSuite) whenAFollowerIsRemovedThenRemovesIt(t *testing.T, factory func() Observer) {
	observer := factory()
	defer observer.Cleanup()
	proxy := mock.NewProxy()
	if err := observer.Observe(proxy); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	if err := observer.UpdateNodes(nodeSpecExamples...); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to update nodes %v", testutils.Failed, err)
	}

	if err := observer.RemoveNodes(nodeSpecExamples[0]); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to remove a follower %v", testutils.Failed, err)
	}

	expected := nodeSpecExamples[1:]
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

	equal := func() bool {
		actual := make([]*node.Spec, len(proxy.Followers))
		i := 0
		for _, v := range proxy.Followers {
			actual[i] = v
			i++
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i].Name < actual[j].Name
		})

		return reflect.DeepEqual(expected, actual)
	}
	testutils.AsyncAssertion.ItMustBeTrue(t, equal)

}
