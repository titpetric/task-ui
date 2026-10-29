//go:build integration

package agent_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

func TestAgent_Shell(t *testing.T) {
	t.Parallel()
	dir, root := realDir(t), realDir(t)
	shell := func(env ...string) (string, error) {
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	out, err := shell("AGENT_TEST_SHELL=--resume x", "PATH="+fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_WORKSPACE_DIR="+root)
	assert.NoError(t, err, out)
	assert.Contains(t, out, "cwd="+dir+"\n", "Claude starts where the run does")
	assert.Contains(t, out, "sandbox=1\n")
	assert.Contains(t, out, "arg=--dangerously-skip-permissions\narg=--append-system-prompt\narg=# task-ui\n", "the context goes to the system prompt")
	assert.Contains(t, out, "arg=--resume\narg=x\n")

	out, err = shell("AGENT_TEST_SHELL=", "PATH="+fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=", "ANTHROPIC_AUTH_TOKEN=", "ANTHROPIC_BASE_URL=")
	assert.Error(t, err)
	assert.Contains(t, out, "CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_AUTH_TOKEN: set up Settings")
}
