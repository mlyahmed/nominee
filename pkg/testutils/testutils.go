package testutils

import (
	"github/mlyahmed.io/nominee/pkg/base"
	"testing"
	"time"
)

type asyncAssertion struct{}

var AsyncAssertion = asyncAssertion{}

const (
	// Succeed ...
	Succeed = "\u2713"

	// Failed ...
	Failed = "\u2717"

	settleTime = 200 * time.Millisecond
	stepTime   = settleTime / 10
)

func (asyncAssertion) ItMustKeepRunning(t *testing.T, c base.DoneChan) {
	t.Helper()
	select {
	case <-c:
		t.Fatalf("\t\t%s FAIL: expected to keep running. But actually not.", Failed)
	default:
		return
	}
}

func (asyncAssertion) ItMustBeStopped(t *testing.T, c base.DoneChan) {
	t.Helper()
	start := time.Now()
	timer := time.NewTimer(stepTime)
	defer timer.Stop()

	for time.Since(start) < settleTime {
		select {
		case <-c:
			return
		case <-timer.C:
			timer.Reset(stepTime)
		}
	}

	t.Fatalf("\t\t%s FAIL: Stonither, expected to be stopped. But actually not.", Failed)
}

func (asyncAssertion) ItMustBeTrue(t *testing.T, assert func() bool) {
	t.Helper()
	start := time.Now()
	timer := time.NewTimer(stepTime)
	defer timer.Stop()

	for time.Since(start) < settleTime {
		if assert() {
			return
		}
		<-timer.C
		timer.Reset(stepTime)
	}

	t.Fatalf("\t\t%spec FATAL: expected to be true. Actually it returns false", Failed)
}
