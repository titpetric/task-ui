package sshagent_test

import (
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshagent"
)

func TestLoad_NoKeys(t *testing.T) {
	t.Parallel()
	a, err := sshagent.Load(t.TempDir(), t.TempDir())
	assert.NoError(t, err)
	assert.Nil(t, a)

	_, err = sshagent.Load(filepath.Join(t.TempDir(), "missing"), t.TempDir())
	assert.Error(t, err)
}

func TestAgent_Close(t *testing.T) {
	t.Parallel()
	var a *sshagent.Agent
	assert.NoError(t, a.Close())
}
