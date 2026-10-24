package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// A run gets HOME at the workspace's home, ATKINS_DB naming the home's
// history and the instance's mirror, and what the instance keeps in the
// server's HOME stays in use: the kept list, and nothing else.
func TestHomeEnv(t *testing.T) {
	t.Parallel()
	environ := []string{"PATH=/bin", "HOME=/root", "GOPATH=/go"}
	env := runner.HomeEnv(environ, "/data/malta/home")
	assert.Equal(t, []string{
		"DOCKER_CONFIG=/root/.docker",
		"CLAUDE_CONFIG_DIR=/root/.claude",
		"XDG_CACHE_HOME=/root/.cache",
		"HOME=/data/malta/home",
		"TASK_UI_HOME=/data/malta/home",
		"ATKINS_DB=/data/malta/home/.atkins/atkins.db:/root/.atkins/atkins.db",
	}, env, "GOPATH, set by the server, stays")
}

// The server's HOME is no fallback: a run with a home gets it even when
// the home is the server's own, and a server without a HOME leaves the
// kept variables alone and the mirror out.
func TestHomeEnv_server(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{
		"DOCKER_CONFIG=/root/.docker",
		"CLAUDE_CONFIG_DIR=/root/.claude",
		"XDG_CACHE_HOME=/root/.cache",
		"GOPATH=/root/go",
		"HOME=/root",
		"TASK_UI_HOME=/root",
		"ATKINS_DB=/root/.atkins/atkins.db:/root/.atkins/atkins.db",
	}, runner.HomeEnv([]string{"HOME=/root"}, "/root"))
	assert.Equal(t, []string{
		"HOME=/data/home",
		"TASK_UI_HOME=/data/home",
		"ATKINS_DB=/data/home/.atkins/atkins.db",
	}, runner.HomeEnv([]string{"PATH=/bin"}, "/data/home"))
	assert.Nil(t, runner.HomeEnv([]string{"HOME=/root"}, ""), "no workspace known")
}
