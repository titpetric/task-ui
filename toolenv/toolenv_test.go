package toolenv

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
)

// stub writes an executable that prints out, whatever the arguments.
func stub(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	assert.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755))
	return path
}

// The detection finds the tools of a PATH, each with its version, its
// path and what it is used for; a tool the PATH does not have is
// absent.
func TestDetect_Tools(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	git := stub(t, dir, "git", `echo "git version 2.43.0"`)
	node := stub(t, dir, "node", `echo "v20.11.0"`)
	docker := stub(t, dir, "docker", `echo "Docker version 25.0.3"`)

	tools := Detect{Path: dir, Timeout: time.Second}.Tools(context.Background())
	byName := map[string]Tool{}
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	assert.Len(t, tools, 4, "git, node and docker, whose compose is there too")
	assert.Equal(t, Tool{Name: "git", Category: CategoryTooling, Version: "2.43.0", Path: git, Description: "version control"}, byName["git"])
	assert.Equal(t, "v20.11.0", byName["node"].Version)
	assert.Equal(t, node, byName["node"].Path)
	assert.Equal(t, CategoryRuntime, byName["node"].Category)
	assert.Equal(t, docker, byName["docker compose"].Path)
	assert.Equal(t, "multi-container applications", byName["docker compose"].Description)
	_, there := byName["go"]
	assert.False(t, there, "a tool the PATH does not have is absent")
}

// A tool whose version command hangs is still there, its version
// unknown, and the detection returns within the timeout.
func TestDetect_Timeout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stub(t, dir, "go", "sleep 5\necho go version go1.22.0")
	start := time.Now()
	tools := Detect{Path: dir, Timeout: 50 * time.Millisecond}.Tools(context.Background())
	assert.Len(t, tools, 1)
	assert.Equal(t, "go", tools[0].Name)
	assert.Equal(t, "", tools[0].Version)
	assert.True(t, time.Since(start) < 2*time.Second, "the detection returns within the timeout")
}

// An empty PATH finds nothing.
func TestDetect_Empty(t *testing.T) {
	t.Parallel()
	assert.Len(t, Detect{Path: t.TempDir()}.Tools(context.Background()), 0)
}

// The tooling reads as TASK_UI_TOOLING, one a line with the name and
// what the tool is used for.
func TestEnv(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", Env(nil))
	assert.Equal(t, "git: version control\ngo: Go toolchain", Env([]Tool{
		{Name: "git", Description: "version control"},
		{Name: "go", Description: "Go toolchain"},
	}))
}

// A spec is found by the URL-safe ID and by the name the page shows.
func TestSpecByID(t *testing.T) {
	t.Parallel()
	spec, ok := SpecByID("docker-compose")
	assert.True(t, ok)
	assert.Equal(t, "docker compose", spec.Name)

	spec, ok = SpecByName("docker compose")
	assert.True(t, ok)
	assert.Equal(t, "docker-compose", spec.ID)

	_, ok = SpecByID("nope")
	assert.False(t, ok)
}

// The version line drops the tool's name and the word version.
func TestVersionLine(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "go1.22 linux/amd64", versionLine("go version go1.22 linux/amd64\n", "go"))
	assert.Equal(t, "2.43.0", versionLine("  git version 2.43.0  \n", "git"))
	assert.Equal(t, "jq-1.7.1", versionLine("jq-1.7.1\n", "jq"))
	assert.Equal(t, "", versionLine("\n", "make"))
}
