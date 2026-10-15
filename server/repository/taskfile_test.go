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
ui:
  allow: [ops]
jobs:
  deploy:
    desc: "Deploy"
    summary: |
      Long text.
    allow: [admin]
    steps:
      - run: echo deploy
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
	if !slices.Equal(project.Allow, []string{"ops"}) {
		t.Fatalf("allow: %v", project.Allow)
	}

	got := []string{}
	for _, r := range project.Runbooks {
		got = append(got, r.Name)
	}
	if want := []string{"deploy", "list", "up"}; !slices.Equal(got, want) {
		t.Fatalf("runbooks: got %v, want %v", got, want)
	}

	deploy := project.Runbook("deploy")
	if deploy.Summary != "Long text." || !slices.Equal(deploy.Allow, []string{"admin"}) {
		t.Fatalf("deploy: %+v", deploy)
	}
	if up := project.Runbook("up"); up.Description != "docker compose up" {
		t.Fatalf("scalar job description: %q", up.Description)
	}
	if list := project.Runbook("list"); list.Description != "echo one" {
		t.Fatalf("list job description: %q", list.Description)
	}
}

func TestLoadProjectTaskfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Taskfile.yml")
	write(t, path, `
version: "3"
tasks:
  build:
    desc: "Build"
    aliases: [b]
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
	for _, r := range project.Runbooks {
		got = append(got, r.Name)
	}
	if want := []string{"build", "shell"}; !slices.Equal(got, want) {
		t.Fatalf("runbooks: got %v, want %v", got, want)
	}
	if project.Runbook("b") == nil || project.Runbook("b").Name != "build" {
		t.Fatal("alias lookup failed")
	}
	if !project.Runbook("shell").Interactive {
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
			{ID: "web", Path: "web site", Title: "Website", Allow: []string{"ops"}, Command: "make"},
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
	if infra.Kind != "atkins" || infra.Runbook("plan") == nil {
		t.Fatalf("infra: %+v", infra)
	}
	web := repo.Find("web")
	if web.Title != "Website" || !slices.Equal(web.Allow, []string{"ops"}) || !slices.Equal(web.Command, []string{"make"}) {
		t.Fatalf("web overrides not applied: %+v", web)
	}
	if sync := repo.Find("sync"); sync == nil || sync.Runbook("sync") == nil {
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
