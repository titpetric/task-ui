package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A finished workflow run lasted from start to finish; one in progress
// so far; one not started, nothing.
func TestWorkflowRun_Elapsed(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 19, 7, 0, 0, 0, time.UTC)
	end := start.Add(95 * time.Second)
	assert.Equal(t, 95*time.Second, (&model.WorkflowRun{StartedAt: &start, FinishedAt: &end}).Elapsed())
	running := time.Now().Add(-time.Minute)
	assert.True(t, (&model.WorkflowRun{StartedAt: &running}).Elapsed() >= time.Minute)
	assert.Equal(t, time.Duration(0), (&model.WorkflowRun{}).Elapsed())
	var none *model.WorkflowRun
	assert.Equal(t, time.Duration(0), none.Elapsed())
}

// The run's answer is the markdown of the last prompt or summary step
// that answered, without the space around it.
func TestWorkflowRun_AnswerFrom(t *testing.T) {
	t.Parallel()
	rows := []model.WorkflowRunStep{
		{Kind: model.StepJob, Folder: "infra", Job: "plan", State: model.StateOK},
		{Kind: model.StepPrompt, Prompt: "fetch the weather", Answer: "rain, 3 mm", State: model.StateOK},
		{Kind: model.StepPrompt, Prompt: "water today?", Answer: "\n**No**, it rained.\n\n", State: model.StateOK},
		{Kind: model.StepJob, Folder: "infra", Job: "apply", State: model.StateOK},
	}
	assert.Equal(t, "**No**, it rained.", replyOf(rows))

	rows[2].Answer, rows[2].State = "", model.StateFailed
	assert.Equal(t, "rain, 3 mm", replyOf(rows), "a prompt without an answer leaves the one before")

	rows = append(rows, model.WorkflowRunStep{Kind: model.StepSummary, Answer: "Step 3 failed.", State: model.StateOK})
	assert.Equal(t, "Step 3 failed.", replyOf(rows))
}

// A workflow of commands alone has no answer: its reply is how the run
// ended.
func TestWorkflowRun_AnswerFrom_commands(t *testing.T) {
	t.Parallel()
	rows := []model.WorkflowRunStep{
		{Kind: model.StepJob, Folder: "infra", Job: "plan", State: model.StateOK},
		{Kind: model.StepWorkflow, WorkflowRef: "w2", State: model.StateOK},
	}
	assert.Equal(t, "", replyOf(rows))
	assert.Equal(t, "", replyOf(nil))
}

// replyOf is what a run takes from its steps, over a stale one.
func replyOf(steps []model.WorkflowRunStep) string {
	run := &model.WorkflowRun{Answer: "stale"}
	run.AnswerFrom(steps)
	return run.Answer
}
