//go:build integration

package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// forgejo-sync takes --dry-run and --no-color alone, and ends with an
// error while the Forgejo integration has no URL and token.
func TestForgejoSync(t *testing.T) {
	t.Setenv("TASK_UI_DATA", t.TempDir())
	t.Setenv("TASK_UI_WORKSPACE", t.TempDir())
	t.Setenv("TASK_UI_RUN_ID", "")
	var out bytes.Buffer
	err := forgejoSync(context.Background(), []string{"--all"}, &out)
	assert.Error(t, err)
	assert.Equal(t, "usage: task-ui forgejo-sync [--dry-run] [--no-color]", err.Error())

	err = forgejoSync(context.Background(), []string{"--dry-run", "--no-color"}, &out)
	assert.ErrorIs(t, err, errForgejoOff)
	assert.Equal(t, "", out.String())
}

// task-ui cleanup in a run of the root's @cleanup works on the data
// directory the run inherits, whether the run has the default workspace,
// TASK_UI_WORKSPACE _ with the root in TASK_UI_WORKSPACE_DIR, or none,
// no TASK_UI_WORKSPACE at all: the run's workspace name is never the
// root. Outside a run the server's TASK_UI_WORKSPACE is the root.
func TestCleanup_InRun(t *testing.T) {
	data := t.TempDir()
	// The server's HOME and a PATH without atkins: the cleanup measures
	// the tool history and prunes none of it.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Setenv("TASK_UI_DATA", data)
	t.Setenv("TASK_UI_RUN_ID", "run")
	t.Setenv("TASK_UI_WORKSPACE", "_")
	t.Setenv("TASK_UI_WORKSPACE_DIR", filepath.Join(data, "workspaces"))
	var out bytes.Buffer
	assert.NoError(t, cleanup(context.Background(), []string{"--dry-run"}, &out))
	assert.Contains(t, out.String(), "task-ui cleanup of "+data+", a dry run: nothing changes")

	t.Setenv("TASK_UI_WORKSPACE", "")
	t.Setenv("TASK_UI_WORKSPACE_DIR", "")
	out.Reset()
	assert.NoError(t, cleanup(context.Background(), []string{"--dry-run"}, &out))
	assert.Contains(t, out.String(), "task-ui cleanup of "+data+", a dry run: nothing changes")

	t.Setenv("TASK_UI_RUN_ID", "")
	t.Setenv("TASK_UI_WORKSPACE", "_")
	err := cleanup(context.Background(), []string{"--dry-run"}, &out)
	assert.Error(t, err)
	assert.Equal(t, "TASK_UI_WORKSPACE: _ is not a directory", err.Error())
}
