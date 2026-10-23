package sshagent_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshagent"
)

func TestAgent_Close(t *testing.T) {
	t.Parallel()
	var a *sshagent.Agent
	assert.NoError(t, a.Close())
}
