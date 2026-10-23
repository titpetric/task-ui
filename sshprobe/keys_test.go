package sshprobe_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// A probe's keys come from Keys, in an agent the probe closes.
func TestKeys(t *testing.T) {
	t.Parallel()
	var keys sshprobe.Keys = &agentKeys{t: t}
	agent, err := keys.LoginAgent(context.Background())
	assert.NoError(t, err)
	assert.Len(t, agent.Keys, 1)
	assert.NoError(t, agent.Close())
}
