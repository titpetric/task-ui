package workflow_test

import (
	"context"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/workflow"
	"github.com/titpetric/task-ui/internal/assert"
)

// claude answers every prompt step with a run and no answer of its own,
// so a check that follows reads the run's output through Output; text is
// what that output says.
type claude struct {
	text string

	mu    sync.Mutex
	asked int
}

func (c *claude) AskStep(ctx context.Context, user *model.User, run *model.WorkflowRun, step *model.WorkflowRunStep, place, input string) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.asked++
	return "run" + string(rune('0'+c.asked)), "", nil
}

func (c *claude) StepText(ctx context.Context, row *model.WorkflowRunStep) string { return c.text }

func (c *claude) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.asked
}

// A check with nothing to look for is refused where the step is
// resolved, before a run of it starts.
func TestRunnerResolveCheck(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := workflow.New(nil, nil, nil, nil)

	steps, err := r.Resolve(ctx, []model.WorkflowStep{{Kind: model.StepCheck, Markers: " Next: , TODO: "}})
	assert.NoError(t, err)
	assert.True(t, steps[0].Check)
	assert.Equal(t, []string{"Next:", "TODO:"}, steps[0].Markers)
	assert.Equal(t, model.CheckPasses, steps[0].Passes)

	_, err = r.Resolve(ctx, []model.WorkflowStep{{Kind: model.StepCheck}})
	assert.ErrorIs(t, err, workflow.ErrCheckMarkers)
}

// Anything that reads a step's output is an Output: the notes board does
// it by replaying the step's recording.
func TestOutput(t *testing.T) {
	t.Parallel()
	var out workflow.Output = &claude{text: "built"}
	assert.Equal(t, "built", out.StepText(context.Background(), &model.WorkflowRunStep{}))
}
