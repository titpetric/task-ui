package folder

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  build:\n    desc: Build the site\n")
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(dir, "scripts", "build.sh"), "# Build with the script\n")
	write(t, filepath.Join(dir, "package.json"), `{"name": "site", "description": "The site"}`)

	// The pipeline file stands for the folder: the Makefile, the scripts
	// and the package scripts beside it are not jobs of their own.
	f := Load("site", dir)
	ids := []string{}
	for _, job := range f.Jobs {
		ids = append(ids, job.ID)
	}
	if want := []string{"build"}; !slices.Equal(ids, want) {
		t.Fatalf("ids: %v", ids)
	}
	if len(f.Sources) != 1 || f.Sources[0].Kind != "task" || f.Description != "" {
		t.Fatalf("pipeline only: %+v", f.Sources)
	}

	// The configuration can still widen a folder to the other conventions.
	wide := Load("site", dir, "task", "make", "script", "npm")
	ids = ids[:0]
	for _, job := range wide.Jobs {
		ids = append(ids, job.ID)
	}
	if want := []string{"task:build", "make:build", "script:build"}; !slices.Equal(ids, want) {
		t.Fatalf("widened ids: %v", ids)
	}
	if wide.Title != "site" || wide.Description != "The site" {
		t.Fatalf("title from the first named source: %q %q", wide.Title, wide.Description)
	}
	if job := wide.Job("make:build"); !slices.Equal(job.Command, []string{"make", "build"}) || job.Key() != "site/make:build" || job.Source != "Makefile" {
		t.Fatalf("make:build: %+v", job)
	}
	if script := wide.Job("script:build"); !slices.Equal(script.Command, []string{"sh", "scripts/build.sh"}) || script.Description != "Build with the script" {
		t.Fatalf("script: %+v", script)
	}

	// Without a pipeline file every convention is a source.
	plain := t.TempDir()
	write(t, filepath.Join(plain, "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(plain, "scripts", "build.sh"), "# Build with the script\n")
	mixed := Load("tools", plain)
	if len(mixed.Sources) != 2 || mixed.Job("make:build") == nil || mixed.Job("script:build") == nil {
		t.Fatalf("mixed: %+v", mixed.Jobs)
	}

	only := Load("site", dir, "make")
	if len(only.Sources) != 1 || only.Jobs[0].ID != "build" || !slices.Equal(only.Jobs[0].Command, []string{"make", "build"}) {
		t.Fatalf("narrowed: %+v", only.Jobs)
	}

	empty := Load("empty", t.TempDir())
	if len(empty.Sources) != 0 || len(empty.Jobs) != 0 || empty.Error != "" {
		t.Fatalf("empty: %+v", empty)
	}
}

func TestLoadErrors(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "atkins.yml"), "jobs: [unclosed")
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	f := Load("bad", dir)
	if f.Error == "" || len(f.Jobs) != 0 || len(f.Sources) != 1 {
		t.Fatalf("a broken pipeline is still the folder's pipeline: %+v", f)
	}
	wide := Load("bad", dir, "atkins", "make")
	if wide.Error == "" || len(wide.Jobs) != 1 || wide.Jobs[0].ID != "make:build" {
		t.Fatalf("broken source keeps the others when widened: %+v", wide)
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "jobs.yml"), "name: Sync\njobs:\n  sync:\n    desc: Sync\n")
	f := LoadFile("sync", filepath.Join(dir, "jobs.yml"))
	if f.Title != "Sync" || f.Dir != dir || f.Job("sync") == nil || f.Job("sync").ID != "sync" {
		t.Fatalf("single file: %+v", f)
	}
	if bad := LoadFile("bad", filepath.Join(dir, "README.md")); bad.Error == "" {
		t.Fatalf("unreadable file: %+v", bad)
	}
}

func TestAppendUnique(t *testing.T) {
	got := appendUnique([]string{"a"}, "b", " a ", "", "b")
	if !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("%v", got)
	}
}
