//go:build integration

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
//
// It runs a process per tool, which is why this is here and not in the unit
// suite: Tools is exec.CommandContext and nothing else.
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
	// lifetime: the stub outlives the 50ms timeout, which is what this
	// asserts: a tool that hangs reports no version.
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
