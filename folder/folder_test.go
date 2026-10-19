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

// The lines of a file the folder's pipeline reads, for its Show button:
// a last line without a newline counts; a file it cannot read is 0.
func TestFolder_Lines(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "atkins.yml"), "jobs:\n  a: echo a\n")
	write(t, filepath.Join(dir, "vars.yml"), "x: 1\ny: 2")
	write(t, filepath.Join(dir, "empty.yml"), "")
	f := &Folder{ID: "f", Dir: dir}

	for name, want := range map[string]int{"atkins.yml": 2, "vars.yml": 2, "empty.yml": 0, "missing.yml": 0, "../escape.yml": 0} {
		if got := f.Lines(name); got != want {
			t.Errorf("%s: %d lines, want %d", name, got, want)
		}
	}
}
