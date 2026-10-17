package detector

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestMakeDetector(t *testing.T) {
	dir := t.TempDir()
	if (makeDetector{}).Detect(dir) != nil {
		t.Fatal("no Makefile, yet a source")
	}
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
	source := (makeDetector{}).Detect(dir)
	if source == nil || source.Error != "" || source.Kind != "make" {
		t.Fatalf("source: %+v", source)
	}
	if !slices.Equal(source.Command, []string{"make"}) {
		t.Fatalf("command: %v", source.Command)
	}
	names := []string{}
	for _, target := range source.Targets {
		names = append(names, target.Name)
	}
	if want := []string{"build", "test", "generate", "lint", "release", "help"}; !slices.Equal(names, want) {
		t.Fatalf("targets: got %v, want %v", names, want)
	}
	build := source.Target("build")
	if build.Description != "compile everything" || !slices.Equal(build.DependsOn, []string{"generate"}) || !slices.Equal(build.Steps, []string{"go build ./..."}) {
		t.Fatalf("build: %+v", build)
	}
	test := source.Target("test")
	if test.Description != "" || !slices.Equal(test.Steps, []string{"go test ./...", "go vet ./..."}) {
		t.Fatalf("test: %+v", test)
	}
	if gen := source.Target("generate"); gen.Description != "Generate code" {
		t.Fatalf("generate: %+v", gen)
	}
	if rel := source.Target("release"); rel.Description != "two targets on one rule" {
		t.Fatalf("release: %+v", rel)
	}

	write(t, filepath.Join(dir, "GNUmakefile"), "all:\n\techo gnu\n")
	if source := (makeDetector{}).Detect(dir); source.File != "GNUmakefile" || !slices.Equal(source.Command, []string{"make", "-f", "GNUmakefile"}) {
		t.Fatalf("GNUmakefile first: %+v", source)
	}
}
