//go:build integration

package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// forgejo-sync takes --dry-run alone, and ends with an error while the
// Forgejo integration has no URL and token.
func TestForgejoSync(t *testing.T) {
	t.Setenv("TASK_UI_DATA", t.TempDir())
	t.Setenv("TASK_UI_WORKSPACE", t.TempDir())
	t.Setenv("TASK_UI_RUN_ID", "")
	var out bytes.Buffer
	err := forgejoSync(context.Background(), []string{"--all"}, &out)
	assert.Error(t, err)
	assert.Equal(t, "usage: task-ui forgejo-sync [--dry-run]", err.Error())

	err = forgejoSync(context.Background(), []string{"--dry-run"}, &out)
	assert.ErrorIs(t, err, errForgejoOff)
	assert.Equal(t, "", out.String())
}
