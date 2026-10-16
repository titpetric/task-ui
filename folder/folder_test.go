package folder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/detector"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFolder_Job(t *testing.T) {
	f := &Folder{ID: "site"}
	f.Sources = []*detector.Source{
		{Kind: "task", File: "Taskfile.yml", Command: []string{"task"}, Targets: []*detector.Target{{Name: "build", Aliases: []string{"b"}}}},
		{Kind: "make", File: "Makefile", Command: []string{"make"}, Targets: []*detector.Target{{Name: "build"}}},
	}
	finish(f)
	if job := f.Job("make:build"); job == nil || job.Kind != "make" {
		t.Fatalf("by id: %+v", job)
	}
	if job := f.Job("build"); job == nil || job.ID != "task:build" {
		t.Fatalf("by name falls back to the first source: %+v", job)
	}
	if job := f.Job("b"); job == nil || job.ID != "task:build" {
		t.Fatalf("by alias: %+v", job)
	}
	if f.Job("deploy") != nil {
		t.Fatal("missing job found")
	}
	var none *Folder
	if none.Job("build") != nil {
		t.Fatal("nil folder returned a job")
	}
}

func TestFolder_SourceJobs(t *testing.T) {
	f := &Folder{ID: "site"}
	f.Sources = []*detector.Source{
		{Kind: "make", File: "Makefile", Targets: []*detector.Target{{Name: "build"}, {Name: "test"}}},
		{Kind: "npm", File: "package.json", Targets: []*detector.Target{{Name: "build"}}},
	}
	finish(f)
	if jobs := f.SourceJobs(f.Sources[0]); len(jobs) != 2 || jobs[1].ID != "make:test" {
		t.Fatalf("make jobs: %+v", jobs)
	}
	if jobs := f.SourceJobs(f.Sources[1]); len(jobs) != 1 || jobs[0].ID != "npm:build" {
		t.Fatalf("npm jobs: %+v", jobs)
	}
}
