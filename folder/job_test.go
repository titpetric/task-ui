package folder

import (
	"testing"
)

func TestJob_Key(t *testing.T) {
	job := &Job{Folder: "infra", ID: "make:build"}
	if job.Key() != "infra/make:build" {
		t.Fatalf("key: %q", job.Key())
	}
}

func TestJob_ArgsCommand(t *testing.T) {
	job := &Job{ID: "@compose:custom", Command: []string{"atkins", "compose:custom"}}
	if got := job.ArgsCommand(); got != "" {
		t.Fatalf("a job without args: %q", got)
	}
	job.TakesArgs = true
	if got := job.ArgsCommand(); got != "atkins compose:custom --" {
		t.Fatalf("args after --: %q", got)
	}
	job.Exec = []string{"docker", "compose"}
	if got := job.ArgsCommand(); got != "docker compose" {
		t.Fatalf("args after the exec step: %q", got)
	}
	if (*Job)(nil).ArgsCommand() != "" {
		t.Fatal("nil job")
	}
}
