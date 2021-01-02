package stonither_test

import (
	"github.com/pkg/errors"
	"github/mlyahmed.io/nominee/infra"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/stonither"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestStonither_when_created_then_it_must_keep_running(t *testing.T) {
	t.Logf("Given the stonither does not exist.")
	{
		t.Logf("\tWhen created.")
		{
			s := stonither.NewBase()
			itMustKeepRunning(t, s.Done())
		}
	}
}

func TestStonither_when_stonith_then_stop(t *testing.T) {
	t.Logf("Given a running stonither")
	{
		s := stonither.NewBase()

		t.Logf("\tWhen it is stonithed")
		{
			s.Stonith()
			itMustBeStopped(t, s.Done())
		}

	}
}

func TestStonither_when_receive_os_signal_then_stop(t *testing.T) {
	t.Logf("Given a running stonither")
	{
		for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
			t.Run(sig.String(), func(t *testing.T) {
				signal.Ignore(stonither.ShutdownSignals...) // So the test it not interrupted
				s := stonither.NewBase()
				t.Logf("\tWhen receive %s signal", sig.String())
				{
					if err := syscall.Kill(syscall.Getpid(), sig); err != nil {
						t.Fatalf("\t\t%s FAIL: Stonither, error when send SIGINT %v", infra.Failed, err)
					}
					itMustBeStopped(t, s.Done())
				}
			})
		}

	}
}

func TestStonither_when_receive_an_error_then_stop(t *testing.T) {
	t.Logf("Given a running stonither")
	{
		s := stonither.NewBase()
		t.Logf("\tWhen receive an error")
		{
			go func() { s.ErrorChan <- errors.New("") }() // So avoid any blocking
			itMustBeStopped(t, s.Done())
		}

	}
}

func TestStonither_when_receive_a_nil_error_then_keep_running(t *testing.T) {
	t.Logf("Given a running stonither")
	{
		s := stonither.NewBase()
		t.Logf("\tWhen receive a nil error")
		{
			go func() { s.ErrorChan <- nil }() // So avoid any blocking

			itMustKeepRunning(t, s.Done())
		}
	}
}

func itMustKeepRunning(t *testing.T, c node.StopChan) {
	select {
	case <-c:
		t.Fatalf("\t\t%s FAIL: Stonither, expected to keep running. But actually not.", infra.Failed)
	default:
		t.Logf("\t\t%s It must keep running.", infra.Succeed)
	}
}

func itMustBeStopped(t *testing.T, c node.StopChan) {
	t.Helper()
	const settleTime = 100 * time.Millisecond
	start := time.Now()
	timer := time.NewTimer(settleTime / 10)
	defer timer.Stop()

	for time.Since(start) < settleTime {
		select {
		case <-c:
			t.Logf("\t\t%s It must be stopped.", infra.Succeed)
			return
		case <-timer.C:
			timer.Reset(settleTime / 10)
		}
	}

	t.Fatalf("\t\t%s FAIL: Stonither, expected to be stopped. But actually not.", infra.Failed)
}
