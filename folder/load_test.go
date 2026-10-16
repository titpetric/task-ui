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

	f := Load("site", dir)
	ids := []string{}
	for _, job := range f.Jobs {
		ids = append(ids, job.ID)
	}
	if want := []string{"task:build", "make:build", "script:build"}; !slices.Equal(ids, want) {
		t.Fatalf("ids: %v", ids)
	}
	if f.Title != "site" || f.Description != "The site" {
		t.Fatalf("title from the first named source: %q %q", f.Title, f.Description)
	}
	if job := f.Job("make:build"); !slices.Equal(job.Command, []string{"make", "build"}) || job.Key() != "site/make:build" || job.Source != "Makefile" {
		t.Fatalf("make:build: %+v", job)
	}
	if script := f.Job("script:build"); !slices.Equal(script.Command, []string{"sh", "scripts/build.sh"}) || script.Description != "Build with the script" {
		t.Fatalf("script: %+v", script)
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
	if f.Error == "" || len(f.Jobs) != 1 || f.Jobs[0].ID != "make:build" {
		t.Fatalf("broken source keeps the others: %+v", f)
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
