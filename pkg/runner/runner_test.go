package runner_test

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	mock2 "github/mlyahmed.io/nominee/pkg/mock"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/runner"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"io/ioutil"
	"syscall"
	"testing"
	"time"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

const settleTime = 100 * time.Millisecond

func TestElectorRunner_when_run_then_keep_running(t *testing.T) {
	t.Logf("Given an ElectorRunner")
	{
		running := false
		r := runner.NewElectorRunner()
		n := mock2.NewNode(t, &node.Spec{})
		e := mock2.NewElector(t)

		t.Logf("\tWhen it is run.")
		{
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				time.Sleep(settleTime)
				running = true
				e.StopChan <- struct{}{}
			}()

			if err := r.Run(ctx, e, n); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}

			if !running {
				t.Fatalf("\t\t%s FAIL: expected to keep running. Actually not.", testutils.Failed)
			}
			t.Logf("\t\t%s It must keep running.", testutils.Succeed)
		}
	}
}

func TestElectorRunner_when_ctx_done_then_stop(t *testing.T) {
	t.Logf("Given an ElectorRunner")
	{
		stopped := false
		r := runner.NewElectorRunner()
		n := mock2.NewNode(t, &node.Spec{})
		e := mock2.NewElector(t)

		t.Logf("\tWhen the context is done.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), settleTime)
			defer cancel()
			go func() {
				<-ctx.Done()
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			if err := r.Run(ctx, e, n); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}
			t.Logf("\t\t%s It must stop.", testutils.Succeed)
			stopped = true
		}
	}
}

func TestElectorRunner_when_elector_stoniths_then_stop(t *testing.T) {
	t.Logf("Given an ElectorRunner")
	{
		stopped := false
		r := runner.NewElectorRunner()
		n := mock2.NewNode(t, &node.Spec{})
		e := mock2.NewElector(t)
		t.Logf("\tWhen the elector stoniths.")
		{
			go func() {
				e.Stonith()
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			if err := r.Run(context.Background(), e, n); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}
			t.Logf("\t\t%s It must stop.", testutils.Succeed)
			stopped = true
		}
	}
}

func TestElectorRunner_when_elector_returns_an_error_then_stop(t *testing.T) {
	t.Logf("Given an ElectorRunner")
	{
		stopped := false
		r := runner.NewElectorRunner()
		n := mock2.NewNode(t, &node.Spec{})
		e := mock2.NewElector(t)

		t.Logf("\tWhen the elector returns an error.")
		{
			err := errors.New("elector")
			e.RunFn = func(node.Node) error {
				return err
			}

			go func() {
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			e := r.Run(context.Background(), e, n)
			if errors.Cause(e) != err {
				t.Fatalf("\t\t%s FAIL: expected to return with the root error. Actually not.", testutils.Failed)
			}
			t.Logf("\t\t%s It must return with the root error.", testutils.Succeed)
			stopped = true
		}
	}
}

func TestObserverRunner_when_run_then_keep_running(t *testing.T) {
	t.Logf("Given an ObserverRunner")
	{
		running := false
		r := runner.NewObserverRunner()
		p := mock2.NewProxy()
		o := mock2.NewObserver(t)

		t.Logf("\tWhen it is run.")
		{
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				time.Sleep(settleTime)
				running = true
				o.Stonith()
			}()

			if err := r.Run(ctx, o, p); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}

			if !running {
				t.Fatalf("\t\t%s FAIL: expected to keep running. Actually not.", testutils.Failed)
			}
			t.Logf("\t\t%s It must keep running.", testutils.Succeed)
		}
	}
}

func TestObserverRunner_when_ctx_done_then_stop(t *testing.T) {
	t.Logf("Given an ObserverRunner")
	{
		stopped := false
		r := runner.NewObserverRunner()
		o := mock2.NewObserver(t)
		p := mock2.NewProxy()

		t.Logf("\tWhen the context is done.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), settleTime)
			defer cancel()
			go func() {
				<-ctx.Done()
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			if err := r.Run(ctx, o, p); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}
			t.Logf("\t\t%s It must stop.", testutils.Succeed)
			stopped = true
		}
	}
}

func TestObserverRunner_when_the_observer_stoniths_then_stop(t *testing.T) {
	t.Logf("Given an ObserverRunner")
	{
		stopped := false
		r := runner.NewObserverRunner()
		o := mock2.NewObserver(t)
		p := mock2.NewProxy()
		t.Logf("\tWhen the observer stoniths.")
		{
			go func() {
				o.Stonith()
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			if err := r.Run(context.Background(), o, p); err != nil {
				t.Fatalf("\t\t%s FAIL: Failed to run: %v", testutils.Failed, err)
			}
			t.Logf("\t\t%s It must stop.", testutils.Succeed)
			stopped = true
		}
	}
}

func TestObserverRunner_when_the_observer_returns_an_error_then_stop(t *testing.T) {
	t.Logf("Given an ObserverRunner")
	{
		stopped := false
		r := runner.NewObserverRunner()
		o := mock2.NewObserver(t)
		p := mock2.NewProxy()

		t.Logf("\tWhen the observer returns an error.")
		{
			err := errors.New("observer")
			o.ObserveFn = func(proxy proxy.Proxy) error {
				return err
			}

			go func() {
				time.Sleep(settleTime)
				if !stopped {
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGABRT)
					t.Fatalf("\t\t%s FAIL: expected to stop. Actually not.", testutils.Failed)
				}
			}()

			e := r.Run(context.Background(), o, p)
			if errors.Cause(e) != err {
				t.Fatalf("\t\t%s FAIL: expected to return with the root error. Actually not.", testutils.Failed)
			}
			t.Logf("\t\t%s It must return with the root error.", testutils.Succeed)
			stopped = true
		}
	}
}
