package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// A run in a workspace with a home project gets HOME there, and what
// the server keeps in its own HOME stays in use.
func TestHomeEnv(t *testing.T) {
	t.Parallel()
	environ := []string{"PATH=/bin", "HOME=/root", "GOPATH=/go"}
	env := runner.HomeEnv(environ, "/data/malta/workspace")
	assert.Equal(t, []string{
		"ATKINS_DB=/root/.atkins/atkins.db",
		"DOCKER_CONFIG=/root/.docker",
		"GIT_CONFIG_GLOBAL=/root/.gitconfig",
		"XDG_CONFIG_HOME=/root/.config",
		"XDG_CACHE_HOME=/root/.cache",
		"XDG_DATA_HOME=/root/.local/share",
		"XDG_STATE_HOME=/root/.local/state",
		"HISTFILE=/root/.bash_history",
		"CLAUDE_CONFIG_DIR=/root/.claude",
		"HOME=/data/malta/workspace",
		"TASK_UI_HOME=/data/malta/workspace",
	}, env, "GOPATH, set by the server, stays")
}

// Without a home project, or with the server's own, the run keeps the
// server's HOME.
func TestHomeEnv_none(t *testing.T) {
	t.Parallel()
	environ := []string{"HOME=/root"}
	assert.Nil(t, runner.HomeEnv(environ, ""))
	assert.Nil(t, runner.HomeEnv(environ, "/root/"))
}
