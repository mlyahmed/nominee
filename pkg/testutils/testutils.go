package testutils

import (
	"github/mlyahmed.io/nominee/pkg/node"
	"testing"
	"time"
)

const (
	// Succeed ...
	Succeed = "\u2713"

	// Failed ...
	Failed = "\u2717"
)

func ItMustKeepRunning(t *testing.T, c node.StopChan) {
	t.Helper()
	select {
	case <-c:
		t.Fatalf("\t\t%s FAIL: expected to keep running. But actually not.", Failed)
	default:
		return //t.Logf("\t\t%s It must keep running.", Succeed)
	}
}

func ItMustBeStopped(t *testing.T, c node.StopChan) {
	t.Helper()
	const settleTime = 100 * time.Millisecond
	start := time.Now()
	timer := time.NewTimer(settleTime / 10)
	defer timer.Stop()

	for time.Since(start) < settleTime {
		select {
		case <-c:
			return
		case <-timer.C:
			timer.Reset(settleTime / 10)
		}
	}

	t.Fatalf("\t\t%s FAIL: Stonither, expected to be stopped. But actually not.", Failed)
}
