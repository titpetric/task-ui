package run

import (
	"testing"
)

func TestRun_Kill(t *testing.T) {
	r := &Run{Key: "k"}
	if r.Killed() {
		t.Fatal("killed before Kill")
	}
	r.Kill()
	if !r.Killed() {
		t.Fatal("not killed after Kill")
	}
	called := 0
	r.SetStop(func() { called++ })
	if called != 1 {
		t.Fatal("a run killed before start must stop as soon as it starts")
	}
	r.Kill()
	if called != 2 {
		t.Fatal("a second Kill did not stop again")
	}
}

func TestRun_SetStop(t *testing.T) {
	r := &Run{Key: "k"}
	called := false
	r.SetStop(func() { called = true })
	if called {
		t.Fatal("stop called without Kill")
	}
	r.Kill()
	if !called {
		t.Fatal("stop not called by Kill")
	}
}

func TestRun_Done(t *testing.T) {
	rs := NewRegistry()
	r := &Run{Key: "k"}
	if _, err := rs.Acquire(r, false); err != nil {
		t.Fatal(err)
	}
	select {
	case <-r.Done():
		t.Fatal("done before release")
	default:
	}
	rs.Release(r)
	<-r.Done()
}
