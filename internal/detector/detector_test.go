package detector

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
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

func TestDetectors(t *testing.T) {
	var kinds []string
	for _, d := range Detectors {
		var _ Detector = d
		kinds = append(kinds, d.Kind())
	}
	if want := []string{"atkins", "task", "make", "npm", "composer", "script"}; !slices.Equal(kinds, want) {
		t.Fatalf("kinds: %v", kinds)
	}
}

func TestByKind(t *testing.T) {
	if d, ok := ByKind("make"); !ok || d.Kind() != "make" {
		t.Fatalf("make: %v %v", d, ok)
	}
	if _, ok := ByKind("ant"); ok {
		t.Fatal("unknown kind found")
	}
}

func TestAny(t *testing.T) {
	dir := t.TempDir()
	if Any(dir) {
		t.Fatal("empty folder has something to run")
	}
	write(t, filepath.Join(dir, "Makefile"), "all:\n\techo\n")
	if !Any(dir) {
		t.Fatal("Makefile not detected")
	}
}

func TestAll(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  build:\n    desc: Build the site\n")
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(dir, "scripts", "build.sh"), "# Build with the script\n")

	kinds := []string{}
	for _, source := range All(dir) {
		kinds = append(kinds, source.Kind)
	}
	if want := []string{"task", "make", "script"}; !slices.Equal(kinds, want) {
		t.Fatalf("sources: %v", kinds)
	}
	only := All(dir, "make")
	if len(only) != 1 || only[0].Kind != "make" {
		t.Fatalf("narrowed: %+v", only)
	}
}

func TestDescribeFromComment(t *testing.T) {
	for line, want := range map[string]string{
		"# Build it":   "Build it",
		"## Build it":  "Build it",
		"#!/bin/bash":  "",
		"#":            "",
		"build: thing": "",
	} {
		got, _ := describeFromComment(line)
		if got != want {
			t.Errorf("%q: got %q, want %q", line, got, want)
		}
	}
}

func TestHasPipeline(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	if HasPipeline(dir) {
		t.Fatal("a Makefile is not a pipeline")
	}
	write(t, filepath.Join(dir, "atkins.yml"), "jobs:\n  build: go build\n")
	if !HasPipeline(dir) {
		t.Fatal("atkins.yml is a pipeline")
	}
	if !Pipeline("task") || Pipeline("make") || !slices.Equal(PipelineKinds, []string{"atkins", "task"}) {
		t.Fatal("pipeline kinds")
	}
}
