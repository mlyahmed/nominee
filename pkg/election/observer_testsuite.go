package election

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"reflect"
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

	testutils.ItMustKeepRunning(t, e.Done())
}

func (observerSuite) whenTheProxyStopsThenStonith(t *testing.T, factory func() Observer) {
	observer := factory()
	defer observer.Cleanup()
	proxy := mock.NewProxy()

	if err := observer.Observe(proxy); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to run %v", testutils.Failed, err)
	}

	proxy.Stonith(context.TODO())

	testutils.ItMustBeStopped(t, observer.Done())
}

func (observerSuite) whenNewLeaderThenPublishIt(t *testing.T, factory func() Observer) {
	for _, leader := range nodeSpecExamples {
		t.Run("", func(t *testing.T) {
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
			testutils.ItMustBeTrue(t, same)
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

	if err := observer.UpdateNodes(nodeSpecExamples); err != nil {
		t.Fatalf("\t\t%s FATAL: Observer, failed to update nodes %v", testutils.Failed, err)
	}

	equal := func() bool {
		actual := make([]*node.Spec, 0)
		for _, v := range proxy.Followers {
			actual = append(actual, v)
		}
		return reflect.DeepEqual(nodeSpecExamples, actual)
	}
	testutils.ItMustBeTrue(t, equal)
}
