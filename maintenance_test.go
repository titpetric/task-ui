package main

import (
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
