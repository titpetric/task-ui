package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
)

// The question of a workflow run's prompt is the prompt, and the input
// under Input: when the run has one.
func TestWorkflowQuestion(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "say ok", note.WorkflowQuestion(" say ok ", " "))
	assert.Equal(t, "say ok\n\nInput:\nloud", note.WorkflowQuestion("say ok", "loud\n"))
}
