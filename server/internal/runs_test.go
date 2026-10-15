package internal

import (
	"errors"
	"testing"
	"time"
)

func TestRunsLock(t *testing.T) {
	runs := NewRuns()

	a := &Run{Key: "infra/deploy", User: "a"}
	if _, err := runs.Acquire(a, false); err != nil {
		t.Fatal(err)
	}

	// Another runbook runs side by side.
	other := &Run{Key: "infra/plan"}
	if _, err := runs.Acquire(other, false); err != nil {
		t.Fatal(err)
	}

	// The same runbook does not.
	current, err := runs.Acquire(&Run{Key: "infra/deploy", User: "b"}, false)
	if !errors.Is(err, ErrBusy) || current != a {
		t.Fatalf("expected busy with run a, got %v %v", current, err)
	}

	// Force kills the active run and waits for it to be released.
	stopped := make(chan struct{})
	a.SetStop(func() {
		close(stopped)
		go func() {
			time.Sleep(20 * time.Millisecond)
			runs.Release(a)
		}()
	})
	b := &Run{Key: "infra/deploy", User: "b"}
	if _, err := runs.Acquire(b, true); err != nil {
		t.Fatal(err)
	}
	<-stopped
	if !a.Killed() || !b.Forced {
		t.Fatal("force did not kill the previous run")
	}
	if got, _ := runs.Get("infra/deploy"); got != b {
		t.Fatal("forced run not registered")
	}

	// Releasing a stale run does not drop the current one.
	runs.Release(a)
	if got, _ := runs.Get("infra/deploy"); got != b {
		t.Fatal("stale release removed the current run")
	}

	runs.Release(b)
	runs.Release(other)
	if len(runs.List()) != 0 {
		t.Fatal("runs left behind")
	}
}

func TestRunsForceTimeout(t *testing.T) {
	runs := NewRuns()
	runs.KillTimeout = 30 * time.Millisecond

	stuck := &Run{Key: "k"}
	runs.Acquire(stuck, false)
	stuck.SetStop(func() {}) // never releases

	if _, err := runs.Acquire(&Run{Key: "k"}, true); !errors.Is(err, ErrBusy) {
		t.Fatalf("expected busy after timeout, got %v", err)
	}
}

func TestKillBeforeStart(t *testing.T) {
	run := &Run{Key: "k"}
	run.Kill()
	called := false
	run.SetStop(func() { called = true })
	if !called {
		t.Fatal("a run killed before start must stop as soon as it starts")
	}
}
