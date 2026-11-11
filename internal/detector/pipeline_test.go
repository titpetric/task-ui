package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPipelineDetectorTaskfile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), `
version: "3"
dotenv: ['.env']
includes:
  docker: ./docker/Taskfile.yml
  docs:
    taskfile: docs/Taskfile.yml
tasks:
  build:
    desc: "Build"
    aliases: [b]
    deps: [lint]
    cmds:
      - go build ./...
      - task: package
  shell:
    desc: "Shell"
    interactive: true
  clean:
    internal: true
    desc: "Clean"
  helper:
    cmds: [echo]
`)
	source := (pipelineDetector{kind: "task", names: taskfileNames}).Detect(dir)
	if source.Kind != "task" || !slices.Equal(source.Command, []string{"task", "-t", "Taskfile.yml"}) {
		t.Fatalf("kind/command: %s %v", source.Kind, source.Command)
	}
	got := []string{}
	for _, target := range source.Targets {
		got = append(got, target.Name)
	}
	if want := []string{"build", "shell"}; !slices.Equal(got, want) {
		t.Fatalf("targets: got %v, want %v", got, want)
	}
	build := source.Target("build")
	if !slices.Equal(build.Steps, []string{"go build ./...", "task: package"}) || !slices.Equal(build.DependsOn, []string{"lint"}) || !slices.Equal(build.Aliases, []string{"b"}) {
		t.Fatalf("taskfile build: %+v", build)
	}
	// The dotenv files a Taskfile names are atkins' and task's to read.
	if data, _ := json.Marshal(source); strings.Contains(string(data), ".env") {
		t.Fatalf("dotenv listed: %s", data)
	}
	if !slices.Equal(source.Includes, []string{"./docker/Taskfile.yml", "docs/Taskfile.yml"}) {
		t.Fatalf("includes: %v", source.Includes)
	}
	if !source.Target("shell").Interactive {
		t.Fatal("interactive flag lost")
	}
}

func TestPipelineDetectorInvalid(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks: [unclosed")
	source := (pipelineDetector{kind: "task", names: taskfileNames}).Detect(dir)
	if source == nil || source.Error == "" {
		t.Fatalf("expected a parse error: %+v", source)
	}
	if (pipelineDetector{kind: "task", names: []string{"Nothing.yml"}}).Detect(dir) != nil {
		t.Fatal("Taskfile detected without one")
	}
}

// A Taskfile task takes no arguments, with any key it carries: the
// command line passing them after -- is an atkins feature.
func TestPipelineDetectorTakesArgs(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  custom:\n    desc: Compose\n    args: true\n    cmds:\n      - docker compose\n")
	task := (pipelineDetector{kind: "task", names: taskfileNames}).Detect(dir)
	if task.Target("custom").TakesArgs {
		t.Fatal("a Taskfile task takes no args")
	}
	if steps := task.Target("custom").Steps; !slices.Equal(steps, []string{"docker compose"}) {
		t.Fatalf("taskfile steps: %q", steps)
	}
}

// isFile tells a regular file from a directory and from nothing.
func TestIsFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n")
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0o755); err != nil {
		t.Fatal(err)
	}

	if !isFile(dir, "Taskfile.yml") {
		t.Fatal("a file is a file")
	}
	if isFile(dir, "folder") {
		t.Fatal("a directory is not a file")
	}
	if isFile(dir, "missing.yml") {
		t.Fatal("nothing is not a file")
	}
}
