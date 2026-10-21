package detector

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPipelineDetectorAtkins(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "atkins.yml"), `
name: Infra
help: Deploys things.
include: vars.yml
env:
  include: [.env, secrets.env]
jobs:
  deploy:
    desc: "Deploy"
    summary: |
      Long text.
    depends_on: build
    env:
      include: deploy.env
    steps:
      - run: echo deploy
      - name: Verify
        run: ./verify.sh
      - task: notify
      - defer: rm -f lock
  build:
    desc: "Build"
    run: go build ./...
  up: docker compose up
  list:
    - echo one
    - echo two
  hidden:
    desc: "Hidden"
    show: false
  nodesc:
    steps:
      - run: echo
  default:
    passthru: true
    steps:
      - task: go:fmt
      - task: go:test
  helper:
    internal: true
    steps:
      - run: echo
`)
	source := (pipelineDetector{kind: "atkins", names: atkinsNames}).Detect(dir)
	if source == nil || source.Error != "" {
		t.Fatalf("source: %+v", source)
	}
	if source.Name != "Infra" || source.Description != "Deploys things." {
		t.Fatalf("name/description: %q %q", source.Name, source.Description)
	}
	if source.Kind != "atkins" || !slices.Equal(source.Command, []string{"atkins", "-f", "atkins.yml"}) {
		t.Fatalf("kind/command: %s %v", source.Kind, source.Command)
	}
	if !slices.Equal(source.Includes, []string{"vars.yml"}) {
		t.Fatalf("includes: %v", source.Includes)
	}
	if !slices.Equal(source.EnvFiles, []string{".env", "secrets.env", "deploy.env"}) {
		t.Fatalf("env files: %v", source.EnvFiles)
	}

	got := []string{}
	for _, target := range source.Targets {
		got = append(got, target.Name)
	}
	// atkins lists every job, as `atkins -l` does: one without a desc takes
	// its first step as its description. Internal and show: false jobs stay out.
	if want := []string{"build", "default", "deploy", "list", "nodesc", "up"}; !slices.Equal(got, want) {
		t.Fatalf("targets: got %v, want %v", got, want)
	}
	if nodesc := source.Target("nodesc"); nodesc.Description != "run: echo" {
		t.Fatalf("nodesc description: %q", nodesc.Description)
	}
	if def := source.Target("default"); def.Description != "task: go:fmt" || !slices.Equal(def.Steps, []string{"task: go:fmt", "task: go:test"}) {
		t.Fatalf("default: %+v", def)
	}

	deploy := source.Target("deploy")
	if deploy.Summary != "Long text." || !slices.Equal(deploy.DependsOn, []string{"build"}) || !slices.Equal(deploy.EnvFiles, []string{"deploy.env"}) {
		t.Fatalf("deploy: %+v", deploy)
	}
	if want := []string{"run: echo deploy", "Verify", "task: notify", "defer: rm -f lock"}; !slices.Equal(deploy.Steps, want) {
		t.Fatalf("deploy steps: got %v, want %v", deploy.Steps, want)
	}
	if build := source.Target("build"); !slices.Equal(build.Steps, []string{"go build ./..."}) {
		t.Fatalf("build steps: %v", build.Steps)
	}
	if up := source.Target("up"); up.Description != "docker compose up" || !slices.Equal(up.Steps, []string{"docker compose up"}) {
		t.Fatalf("scalar job: %+v", up)
	}
	if list := source.Target("list"); list.Description != "echo one" {
		t.Fatalf("list job description: %q", list.Description)
	}
}

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
	if !slices.Equal(source.EnvFiles, []string{".env"}) {
		t.Fatalf("dotenv: %v", source.EnvFiles)
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
	write(t, filepath.Join(dir, "atkins.yml"), "jobs: [unclosed")
	source := (pipelineDetector{kind: "atkins", names: atkinsNames}).Detect(dir)
	if source == nil || source.Error == "" {
		t.Fatalf("expected a parse error: %+v", source)
	}
	if (pipelineDetector{kind: "task", names: taskfileNames}).Detect(dir) != nil {
		t.Fatal("Taskfile detected without one")
	}
}

// An atkins job with `args: true` takes arguments; a Taskfile task
// with the same key does not, and neither does a job without it.
func TestPipelineDetectorTakesArgs(t *testing.T) {
	dir := t.TempDir()
	body := "jobs:\n  custom:\n    desc: Compose\n    args: true\n    steps:\n      - run: docker compose\n  up:\n    desc: Up\n    steps:\n      - run: docker compose up\n"
	write(t, filepath.Join(dir, "atkins.yml"), body)
	write(t, filepath.Join(dir, "Taskfile.yml"), strings.ReplaceAll(body, "jobs:", "tasks:"))
	atkins := (pipelineDetector{kind: "atkins", names: atkinsNames}).Detect(dir)
	if !atkins.Target("custom").TakesArgs || atkins.Target("up").TakesArgs {
		t.Fatalf("atkins args: %+v", atkins.Targets)
	}
	task := (pipelineDetector{kind: "task", names: taskfileNames}).Detect(dir)
	if task.Target("custom").TakesArgs {
		t.Fatal("a Taskfile task takes no args")
	}
}
