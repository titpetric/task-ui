package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/workflow"
)

// A check step reads the step it follows through the board, which is
// the runner's Output: a prompt step's answer comes off the row, and a
// job step whose run is gone, or an instance with no runs at all, reads
// as nothing rather than failing the check.
func TestBoard_StepText(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var out workflow.Output = &note.Board{}
	assert.Equal(t, "Next: the second half", out.StepText(ctx, &model.WorkflowRunStep{Answer: "Next: the second half"}))
	assert.Equal(t, "", out.StepText(ctx, &model.WorkflowRunStep{RunID: "r1"}))
	assert.Equal(t, "", out.StepText(ctx, &model.WorkflowRunStep{}))
}

// The question of a workflow run's prompt is the prompt, and the input
// under Input: when the run has one.
func TestWorkflowQuestion(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "say ok", note.WorkflowQuestion(" say ok ", " "))
	assert.Equal(t, "say ok\n\nInput:\nloud", note.WorkflowQuestion("say ok", "loud\n"))
}
