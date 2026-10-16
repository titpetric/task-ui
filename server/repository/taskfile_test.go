package repository

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

func TestLoadProjectAtkins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "atkins.yml")
	write(t, path, `
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
`)
	project := LoadProject("infra", path)
	if project.Error != "" {
		t.Fatal(project.Error)
	}
	if project.Title != "Infra" || project.Description != "Deploys things." {
		t.Fatalf("title/description: %q %q", project.Title, project.Description)
	}
	if project.Kind != "atkins" || !slices.Equal(project.Command, []string{"atkins", "-f", "atkins.yml"}) {
		t.Fatalf("kind/command: %s %v", project.Kind, project.Command)
	}
	if !slices.Equal(project.Includes, []string{"vars.yml"}) {
		t.Fatalf("includes: %v", project.Includes)
	}
	if !slices.Equal(project.EnvFiles, []string{".env", "secrets.env", "deploy.env"}) {
		t.Fatalf("env files: %v", project.EnvFiles)
	}

	got := []string{}
	for _, r := range project.Jobs {
		got = append(got, r.Name)
	}
	if want := []string{"build", "deploy", "list", "up"}; !slices.Equal(got, want) {
		t.Fatalf("jobs: got %v, want %v", got, want)
	}

	deploy := project.Job("deploy")
	if deploy.Summary != "Long text." || !slices.Equal(deploy.DependsOn, []string{"build"}) || !slices.Equal(deploy.EnvFiles, []string{"deploy.env"}) {
		t.Fatalf("deploy: %+v", deploy)
	}
	if want := []string{"run: echo deploy", "Verify", "task: notify", "defer: rm -f lock"}; !slices.Equal(deploy.Steps, want) {
		t.Fatalf("deploy steps: got %v, want %v", deploy.Steps, want)
	}
	if build := project.Job("build"); !slices.Equal(build.Steps, []string{"go build ./..."}) {
		t.Fatalf("build steps: %v", build.Steps)
	}
	if up := project.Job("up"); up.Description != "docker compose up" || !slices.Equal(up.Steps, []string{"docker compose up"}) {
		t.Fatalf("scalar job: %+v", up)
	}
	if list := project.Job("list"); list.Description != "echo one" {
		t.Fatalf("list job description: %q", list.Description)
	}
}

func TestLoadProjectTaskfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Taskfile.yml")
	write(t, path, `
version: "3"
dotenv: ['.env']
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
	project := LoadProject("web", path)
	if project.Kind != "task" || !slices.Equal(project.Command, []string{"task", "-t", "Taskfile.yml"}) {
		t.Fatalf("kind/command: %s %v", project.Kind, project.Command)
	}
	got := []string{}
	for _, r := range project.Jobs {
		got = append(got, r.Name)
	}
	if want := []string{"build", "shell"}; !slices.Equal(got, want) {
		t.Fatalf("jobs: got %v, want %v", got, want)
	}
	if project.Job("b") == nil || project.Job("b").Name != "build" {
		t.Fatal("alias lookup failed")
	}
	if build := project.Job("build"); !slices.Equal(build.Steps, []string{"go build ./...", "task: package"}) || !slices.Equal(build.DependsOn, []string{"lint"}) {
		t.Fatalf("taskfile build: %+v", build)
	}
	if !slices.Equal(project.EnvFiles, []string{".env"}) {
		t.Fatalf("dotenv: %v", project.EnvFiles)
	}
	if !project.Job("shell").Interactive {
		t.Fatal("interactive flag lost")
	}
	if project.Title != "web" {
		t.Fatalf("title: %q", project.Title)
	}
}

func TestLoadProjectInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "atkins.yml")
	write(t, path, "jobs: [unclosed")
	project := LoadProject("bad", path)
	if project.Error == "" {
		t.Fatal("expected a parse error")
	}
}

func TestProjectsDiscovery(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "Taskfile.yml"), "tasks:\n  root:\n    desc: Root\n")
	write(t, filepath.Join(root, "infra", "atkins.yml"), "name: Infra\njobs:\n  plan: echo plan\n")
	// Atkins files win over Taskfiles in the same folder.
	write(t, filepath.Join(root, "infra", "Taskfile.yml"), "tasks:\n  other:\n    desc: Other\n")
	write(t, filepath.Join(root, "web site", "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	write(t, filepath.Join(root, ".hidden", "Taskfile.yml"), "tasks:\n  x:\n    desc: X\n")
	write(t, filepath.Join(root, "empty", "README.md"), "nothing")
	write(t, filepath.Join(root, "elsewhere", "jobs.yml"), "jobs:\n  sync:\n    desc: Sync\n")

	repo := &Projects{
		Root:     root,
		Discover: true,
		Config: []ProjectConfig{
			{ID: "web", Path: "web site", Title: "Website", Allow: []string{"ops"}, Command: "make", Jobs: map[string][]string{"build": {"admin"}}},
			{ID: "sync", Path: "elsewhere/jobs.yml"},
		},
	}
	list := repo.List()

	ids := []string{}
	for _, p := range list {
		ids = append(ids, p.ID)
	}
	slices.Sort(ids)
	if want := []string{filepath.Base(root), "infra", "sync", "web"}; !slices.Equal(ids, want) {
		t.Fatalf("projects: got %v, want %v", ids, want)
	}

	infra := repo.Find("infra")
	if infra.Kind != "atkins" || infra.Job("plan") == nil {
		t.Fatalf("infra: %+v", infra)
	}
	web := repo.Find("web")
	if web.Title != "Website" || !slices.Equal(web.Allow, []string{"ops"}) || !slices.Equal(web.Command, []string{"make"}) {
		t.Fatalf("web overrides not applied: %+v", web)
	}
	if !slices.Equal(web.Job("build").Allow, []string{"admin"}) {
		t.Fatalf("job allow not applied: %+v", web.Job("build"))
	}
	if sync := repo.Find("sync"); sync == nil || sync.Job("sync") == nil {
		t.Fatalf("configured file project missing: %+v", sync)
	}
}

func TestSafeID(t *testing.T) {
	for in, want := range map[string]string{
		"web site": "web-site",
		"infra":    "infra",
		"../x":     "x",
		"":         "default",
		"a/b c":    "a-b-c",
	} {
		if got := SafeID(in); got != want {
			t.Errorf("SafeID(%q) = %q, want %q", in, got, want)
		}
	}
}
