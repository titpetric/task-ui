package workflow_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/feature/workflow"
	"github.com/titpetric/task-ui/internal/assert"
)

// The notes board runs a workflow's prompt.
func TestPrompter(t *testing.T) {
	t.Parallel()
	var p workflow.Prompter = &note.Board{}
	assert.True(t, p != nil)
}
