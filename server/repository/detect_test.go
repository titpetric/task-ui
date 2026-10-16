package repository

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestMakeDetector(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Makefile"), `
SHELL := /bin/bash
VERSION ?= dev
.PHONY: build test
.DEFAULT_GOAL := help

# Build the binary
build: generate ## compile everything
	go build ./...

test: build
	@go test ./...
	go vet ./...

## Generate code
generate:
	go generate ./...

%.o: %.c
	cc -c $<

$(BIN)/tool:
	echo computed

lint release: ## two targets on one rule
	echo lint

help:
	@grep -E '^[a-z]+:.*##' Makefile
`)
	source := makeDetector{}.Detect(dir)
	if source == nil || source.Error != "" {
		t.Fatalf("source: %+v", source)
	}
	if !slices.Equal(source.Command, []string{"make"}) {
		t.Fatalf("command: %v", source.Command)
	}
	names := []string{}
	for _, job := range source.Jobs {
		names = append(names, job.Name)
	}
	if want := []string{"build", "test", "generate", "lint", "release", "help"}; !slices.Equal(names, want) {
		t.Fatalf("targets: got %v, want %v", names, want)
	}
	byName := map[string]int{}
	for i, job := range source.Jobs {
		byName[job.Name] = i
	}
	build := source.Jobs[byName["build"]]
	if build.Description != "compile everything" || !slices.Equal(build.DependsOn, []string{"generate"}) || !slices.Equal(build.Steps, []string{"go build ./..."}) {
		t.Fatalf("build: %+v", build)
	}
	test := source.Jobs[byName["test"]]
	if test.Description != "" || !slices.Equal(test.Steps, []string{"go test ./...", "go vet ./..."}) {
		t.Fatalf("test: %+v", test)
	}
	if gen := source.Jobs[byName["generate"]]; gen.Description != "Generate code" {
		t.Fatalf("generate: %+v", gen)
	}
	if rel := source.Jobs[byName["release"]]; rel.Description != "two targets on one rule" {
		t.Fatalf("release: %+v", rel)
	}

	write(t, filepath.Join(dir, "GNUmakefile"), "all:\n\techo gnu\n")
	if source := (makeDetector{}).Detect(dir); source.File != "GNUmakefile" || !slices.Equal(source.Command, []string{"make", "-f", "GNUmakefile"}) {
		t.Fatalf("GNUmakefile first: %+v", source)
	}
}

func TestNpmAndComposerDetectors(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "package.json"), `{
  "name": "site", "description": "The marketing site",
  "scripts": {"build": "vite build", "prebuild": "rimraf dist", "dev": "vite", "postinstall": "husky"}
}`)
	source := npmDetector{}.Detect(dir)
	if source.Name != "site" || source.Description != "The marketing site" || !slices.Equal(source.Command, []string{"npm", "run"}) {
		t.Fatalf("npm source: %+v", source)
	}
	names := []string{}
	for _, job := range source.Jobs {
		names = append(names, job.Name)
	}
	if want := []string{"build", "dev", "postinstall"}; !slices.Equal(names, want) {
		t.Fatalf("npm scripts: got %v, want %v", names, want)
	}
	if source.Jobs[0].Description != "vite build" {
		t.Fatalf("npm desc: %+v", source.Jobs[0])
	}
	write(t, filepath.Join(dir, "pnpm-lock.yaml"), "")
	if source := (npmDetector{}).Detect(dir); !slices.Equal(source.Command, []string{"pnpm", "run"}) {
		t.Fatalf("lock file manager: %v", source.Command)
	}

	write(t, filepath.Join(dir, "composer.json"), `{
  "name": "acme/app",
  "scripts": {"test": "phpunit", "lint": ["phpcs", "phpstan"], "post-install-cmd": "echo no"},
  "scripts-descriptions": {"test": "Run the test suite"}
}`)
	composer := composerDetector{}.Detect(dir)
	if composer.Name != "acme/app" || len(composer.Jobs) != 2 {
		t.Fatalf("composer: %+v", composer)
	}
	if composer.Jobs[0].Name != "lint" || composer.Jobs[0].Description != "phpcs" || len(composer.Jobs[0].Steps) != 2 {
		t.Fatalf("composer lint: %+v", composer.Jobs[0])
	}
	if composer.Jobs[1].Description != "Run the test suite" {
		t.Fatalf("composer test: %+v", composer.Jobs[1])
	}
	if _, err := os.Stat(filepath.Join(dir, "composer.json")); err != nil {
		t.Fatal(err)
	}
}

func TestScriptDetector(t *testing.T) {
	dir := t.TempDir()
	if (scriptDetector{}).Detect(dir) != nil {
		t.Fatal("no scripts folder, yet a source")
	}
	write(t, filepath.Join(dir, "scripts", "deploy.sh"), "#!/usr/bin/env bash\n# shellcheck disable=SC2086\n# Deploy the current release\nset -e\n")
	write(t, filepath.Join(dir, "scripts", "tool"), "#!/usr/bin/env python3\n\"\"\"not a comment\"\"\"\n")
	write(t, filepath.Join(dir, "scripts", "README.md"), "# docs\n")
	write(t, filepath.Join(dir, "scripts", ".hidden.sh"), "")
	os.Chmod(filepath.Join(dir, "scripts", "tool"), 0o755)

	source := scriptDetector{}.Detect(dir)
	if source == nil || len(source.Jobs) != 2 {
		t.Fatalf("scripts: %+v", source)
	}
	deploy, tool := source.Jobs[0], source.Jobs[1]
	if deploy.Name != "deploy" || deploy.Description != "Deploy the current release" || !slices.Equal(deploy.Args, []string{"sh", "scripts/deploy.sh"}) {
		t.Fatalf("deploy: %+v", deploy)
	}
	if tool.Name != "tool" || tool.Description != "" || !slices.Equal(tool.Args, []string{"./scripts/tool"}) {
		t.Fatalf("tool: %+v", tool)
	}
}

func TestLoadFolderSeveralSources(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  build:\n    desc: Build the site\n")
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(dir, "scripts", "build.sh"), "# Build with the script\n")

	project := LoadFolder("site", dir)
	kinds := []string{}
	for _, source := range project.Sources {
		kinds = append(kinds, source.Kind)
	}
	if want := []string{"task", "make", "script"}; !slices.Equal(kinds, want) {
		t.Fatalf("sources: %v", kinds)
	}
	ids := []string{}
	for _, job := range project.Jobs {
		ids = append(ids, job.ID)
	}
	if want := []string{"task:build", "make:build", "script:build"}; !slices.Equal(ids, want) {
		t.Fatalf("ids: %v", ids)
	}
	if job := project.Job("make:build"); job == nil || !slices.Equal(job.Command, []string{"make", "build"}) || job.Key() != "site/make:build" {
		t.Fatalf("make:build: %+v", job)
	}
	if job := project.Job("build"); job == nil || job.ID != "task:build" {
		t.Fatalf("name lookup falls back to the first source: %+v", job)
	}
	if script := project.Job("script:build"); !slices.Equal(script.Command, []string{"sh", "scripts/build.sh"}) || script.Description != "Build with the script" {
		t.Fatalf("script: %+v", script)
	}

	only := LoadFolder("site", dir, "make")
	if len(only.Sources) != 1 || only.Jobs[0].ID != "build" {
		t.Fatalf("narrowed: %+v", only.Jobs)
	}
	if !DetectAny(dir) || DetectAny(t.TempDir()) {
		t.Fatal("DetectAny")
	}
}
