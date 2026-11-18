//go:build integration

package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// fakeClaude writes a claude script that runs body, and returns the directory holding it.
func fakeClaude(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "claude")
	assert.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755))
	return dir
}

// claudeAgent returns an agent that runs the fake claude in path with a session config directory.
func claudeAgent(path, config string) *Agent {
	return &Agent{
		Env: []string{
			"PATH=" + path,
			"CLAUDE_CODE_OAUTH_TOKEN=test",
			"TASK_UI_CLAUDE_SESSION=01arz3ndektsv4rrffq69g5fav",
			"TASK_UI_CLAUDE_CONFIG=" + config,
		},
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
}

// A run that ends of its own takes its session directory with it.
func TestAgent_askCleanExit(t *testing.T) {
	t.Parallel()
	path := fakeClaude(t, `echo '{"type":"result","result":"done"}'`)
	config := filepath.Join(t.TempDir(), "ana_01arz3ndektsv4rrffq69g5fav")
	assert.NoError(t, claudeAgent(path, config).ask(context.Background(), t.TempDir(), "hello", "", io.Discard))
	_, err := os.Stat(config)
	assert.True(t, os.IsNotExist(err))
}

// A run killed by a signal keeps its session directory for the restart to resume.
func TestAgent_askKilledKeeps(t *testing.T) {
	t.Parallel()
	path := fakeClaude(t, "kill -TERM $$")
	config := filepath.Join(t.TempDir(), "ana_01arz3ndektsv4rrffq69g5fav")
	err := claudeAgent(path, config).ask(context.Background(), t.TempDir(), "hello", "", io.Discard)
	assert.Error(t, err)
	assert.True(t, killed(err))
	_, statErr := os.Stat(config)
	assert.NoError(t, statErr)
	assert.False(t, killed(nil))
}
