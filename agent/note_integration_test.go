//go:build integration

package agent

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A run without a thread is no note's, and starts nothing.
func TestAgent_Note(t *testing.T) {
	t.Parallel()
	err := (&Agent{}).Note(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TASK_UI_NOTE")
}
