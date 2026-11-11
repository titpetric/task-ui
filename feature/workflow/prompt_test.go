package workflow_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/workflow"
)

// The notes board runs a workflow's prompt.
func TestPrompter(t *testing.T) {
	t.Parallel()
	var p workflow.Prompter = &note.Board{}
	assert.True(t, p != nil)
}
