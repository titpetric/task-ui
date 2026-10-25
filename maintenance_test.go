package main

import (
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The maintenance commands take --dry-run, -n and --, and nothing else.
func TestDryRun(t *testing.T) {
	t.Parallel()
	dry, err := dryRun(nil, "usage")
	assert.NoError(t, err)
	assert.False(t, dry)
	dry, err = dryRun([]string{"--", "--dry-run"}, "usage")
	assert.NoError(t, err)
	assert.True(t, dry)
	dry, err = dryRun([]string{"-n"}, "usage")
	assert.NoError(t, err)
	assert.True(t, dry)
	_, err = dryRun([]string{"now"}, "usage: task-ui cleanup [--dry-run]")
	assert.Error(t, err)
	assert.Equal(t, "usage: task-ui cleanup [--dry-run]", err.Error())
}

// A maintenance command in a run of task-ui, where TASK_UI_WORKSPACE names
// the run's workspace, reads neither it nor TASK_UI_WORKSPACE_DIR: its root
// is the data directory's, <data>/workspaces, with or without a workspace.
// Outside a run it reads the server's configuration, TASK_UI_WORKSPACE the root.
func TestMaintenanceConfig(t *testing.T) {
	data := t.TempDir()
	t.Setenv("TASK_UI_DATA", data)
	t.Setenv("TASK_UI_RUN_ID", "run")
	t.Setenv("TASK_UI_WORKSPACE", "_")
	t.Setenv("TASK_UI_WORKSPACE_DIR", t.TempDir())
	cfg, err := maintenanceConfig()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(data, "workspaces"), cfg.Root)

	t.Setenv("TASK_UI_WORKSPACE", "")
	t.Setenv("TASK_UI_WORKSPACE_DIR", "")
	cfg, err = maintenanceConfig()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(data, "workspaces"), cfg.Root)

	t.Setenv("TASK_UI_RUN_ID", "")
	t.Setenv("TASK_UI_WORKSPACE", "_")
	_, err = maintenanceConfig()
	assert.Error(t, err)
	assert.Equal(t, "TASK_UI_WORKSPACE: _ is not a directory", err.Error())

	server := t.TempDir()
	t.Setenv("TASK_UI_WORKSPACE", server)
	cfg, err = maintenanceConfig()
	assert.NoError(t, err)
	assert.Equal(t, server, cfg.Root)
}
