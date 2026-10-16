package run

import (
	"errors"
	"testing"
	"time"
)

func TestRegistry_Acquire(t *testing.T) {
	rs := NewRegistry()

	a := &Run{Key: "infra/deploy", User: "a"}
	if _, err := rs.Acquire(a, false); err != nil {
		t.Fatal(err)
	}

	// Another job runs side by side.
	other := &Run{Key: "infra/plan"}
	if _, err := rs.Acquire(other, false); err != nil {
		t.Fatal(err)
	}

	// The same job does not.
	current, err := rs.Acquire(&Run{Key: "infra/deploy", User: "b"}, false)
	if !errors.Is(err, ErrBusy) || current != a {
		t.Fatalf("expected busy with run a, got %v %v", current, err)
	}

	// Force kills the active run and waits for it to be released.
	stopped := make(chan struct{})
	a.SetStop(func() {
		close(stopped)
		go func() {
			time.Sleep(20 * time.Millisecond)
			rs.Release(a)
		}()
	})
	b := &Run{Key: "infra/deploy", User: "b"}
	if _, err := rs.Acquire(b, true); err != nil {
		t.Fatal(err)
	}
	<-stopped
	if !a.Killed() || !b.Forced {
		t.Fatal("force did not kill the previous run")
	}
	if got, _ := rs.Get("infra/deploy"); got != b {
		t.Fatal("forced run not registered")
	}

	// Releasing a stale run does not drop the current one.
	rs.Release(a)
	if got, _ := rs.Get("infra/deploy"); got != b {
		t.Fatal("stale release removed the current run")
	}

	if list := rs.List(); len(list) != 2 {
		t.Fatalf("list: %+v", list)
	}
	rs.Release(b)
	rs.Release(other)
	if len(rs.List()) != 0 {
		t.Fatal("runs left behind")
	}
	if _, ok := rs.Get("infra/deploy"); ok {
		t.Fatal("released run still found")
	}
}

func TestRegistry_ForceTimeout(t *testing.T) {
	rs := NewRegistry()
	rs.KillTimeout = 30 * time.Millisecond

	stuck := &Run{Key: "k"}
	if _, err := rs.Acquire(stuck, false); err != nil {
		t.Fatal(err)
	}
	stuck.SetStop(func() {}) // never releases

	if _, err := rs.Acquire(&Run{Key: "k"}, true); !errors.Is(err, ErrBusy) {
		t.Fatalf("expected busy after timeout, got %v", err)
	}
}
