package stonither_test

import (
	"github.com/pkg/errors"
	"github/mlyahmed.io/nominee/pkg/stonither"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"os/signal"
	"syscall"
	"testing"
)

func TestStonither_when_created_then_it_must_keep_running(t *testing.T) {
	t.Logf("Given the stonither does not exist.")
	{
		t.Logf("\tWhen created.")
		{
			s := stonither.NewBase()
			testutils.ItMustKeepRunning(t, s.Done())
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
			testutils.ItMustBeStopped(t, s.Done())
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
						t.Fatalf("\t\t%s FAIL: Stonither, error when send SIGINT %v", testutils.Failed, err)
					}
					testutils.ItMustBeStopped(t, s.Done())
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
			testutils.ItMustBeStopped(t, s.Done())
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

			testutils.ItMustKeepRunning(t, s.Done())
		}
	}
}
