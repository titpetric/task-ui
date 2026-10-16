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
