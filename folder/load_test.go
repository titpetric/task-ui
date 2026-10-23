package folder

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/skills"
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

	// One convention reads a folder, the first there is: atkins, then a
	// Taskfile, a Makefile, package scripts, and the scripts folder last.
	plain := t.TempDir()
	write(t, filepath.Join(plain, "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(plain, "package.json"), `{"scripts":{"dev":"vite"}}`)
	write(t, filepath.Join(plain, "scripts", "build.sh"), "# Build with the script\n")
	mixed := Load("tools", plain)
	if len(mixed.Sources) != 1 || mixed.Sources[0].Kind != "make" || mixed.Job("build") == nil || mixed.Job("dev") != nil {
		t.Fatalf("make before package scripts and scripts: %+v", mixed.Jobs)
	}
	write(t, filepath.Join(plain, "Taskfile.yml"), "tasks:\n  lint:\n    desc: Lint\n")
	if task := Load("tools", plain); len(task.Sources) != 1 || task.Sources[0].Kind != "task" {
		t.Fatalf("a Taskfile before the Makefile: %+v", task.Sources)
	}
	write(t, filepath.Join(plain, "atkins.yml"), "jobs:\n  ship: echo ship\n")
	if atkins := Load("tools", plain); len(atkins.Sources) != 1 || atkins.Sources[0].Kind != "atkins" || atkins.Job("ship") == nil {
		t.Fatalf("atkins first and alone: %+v", atkins.Sources)
	}
	scripts := t.TempDir()
	write(t, filepath.Join(scripts, "scripts", "build.sh"), "# Build with the script\n")
	if s := Load("bin", scripts); len(s.Sources) != 1 || s.Sources[0].Kind != "script" || s.Job("build") == nil {
		t.Fatalf("the scripts folder when there is nothing else: %+v", s.Sources)
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

// A scripts folder with a pipeline lists its loose scripts beside the
// pipeline's jobs, every job under its own name.
func TestLoad_LooseScripts(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "scripts")
	write(t, filepath.Join(dir, "atkins.yml"), "tasks:\n  weather:\n    desc: \"Weather\"\n    cmds:\n      - echo sunny\n")
	write(t, filepath.Join(dir, "bash.sh"), "#!/bin/bash\n# A shell\n# interactive: true\nexec bash\n")
	if err := os.Chmod(filepath.Join(dir, "bash.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	f := Load("deploy/scripts", dir)
	if len(f.Sources) != 2 {
		t.Fatalf("sources: %d", len(f.Sources))
	}
	weather, bash := f.Job("weather"), f.Job("bash")
	if weather == nil || bash == nil {
		t.Fatalf("jobs: %+v", f.Jobs)
	}
	if !bash.Interactive || bash.Kind != "script" || bash.Command[0] != "./bash.sh" {
		t.Fatalf("bash: %+v", bash)
	}
}

// default is the first job of a folder, wherever its file has it.
func TestLoad_DefaultFirst(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n  test:\n    desc: Test\n  default:\n    desc: Vet\n")
	ids := []string{}
	for _, job := range Load("site", dir).Jobs {
		ids = append(ids, job.ID)
	}
	if len(ids) != 3 || ids[0] != "default" {
		t.Fatalf("ids: %v", ids)
	}
}

// A folder takes the lint findings of its sources, in source order.
func TestFinish_Errors(t *testing.T) {
	f := &Folder{ID: "app", Title: "app", Sources: []*detector.Source{
		{Kind: "atkins", File: "atkins.yml", Error: "atkins.yml: pipeline 'app' has errors", Errors: []skills.Finding{{Path: "atkins.yml", Job: "build", Detail: "no go:fmt"}}},
		{Kind: "script", File: "scripts", Aside: true},
	}}
	finish(f)
	if len(f.Errors) != 1 || f.Errors[0].Job != "build" || f.Error != "atkins.yml: pipeline 'app' has errors" {
		t.Fatalf("errors: %+v %q", f.Errors, f.Error)
	}
}
