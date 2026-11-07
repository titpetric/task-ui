//go:build integration

package git

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGit_Collisions(t *testing.T) {
	l := newLab(t)
	l.commit("c.txt", "theirs\n", "adds c")
	l.run(l.work, "fetch", "origin")
	l.save(l.work, "c.txt", "mine\n")
	l.save(l.work, "own.txt", "mine\n")
	if err := os.MkdirAll(filepath.Join(l.work, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	l.save(l.work, "sub/deep.txt", "mine\n")
	out := l.run(l.work, append([]string{"-c", "core.quotePath=false"}, "status", "--porcelain=v2", "--branch")...)
	found := l.git.collisions(context.Background(), l.work, "@{u}", ParseStatus(out))
	if !slices.Equal(found, []string{"c.txt"}) {
		t.Errorf("the collisions are %v, want the file the upstream holds too", found)
	}
}

func TestGit_ParkWork(t *testing.T) {
	l := newLab(t)
	l.save(l.work, "a.txt", "mine\n")
	rec := Recovery{}
	if err := l.git.parkWork(context.Background(), l.work, "20260101-120000.000", "@{u}", &rec, io.Discard); err != nil {
		t.Fatal(err)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "show", rec.Parked[0]+":a.txt"); got != "mine" {
		t.Errorf("the parked ref holds %q, want the change", got)
	}
	if got := l.files(l.work)["a.txt"]; got != "mine\n" {
		t.Errorf("parking changed the file to %q", got)
	}
	clean := Recovery{}
	l.save(l.work, "a.txt", "one\n")
	if err := l.git.parkWork(context.Background(), l.work, "20260101-120001.000", "@{u}", &clean, io.Discard); err != nil {
		t.Fatal(err)
	}
	if len(clean.Parked) != 0 {
		t.Errorf("parked %v of a checkout with nothing at risk", clean.Parked)
	}
}
