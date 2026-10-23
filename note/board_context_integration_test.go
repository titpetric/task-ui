//go:build integration

package note_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A workflow's prompt gets the Workflow context first: the workflow and
// its link, the run's number, starter and input, and the replies of the
// last three finished runs, newest first, without the runs that have
// none and without the run itself, each cut to 1.5 KiB. The first run
// reads that there are no earlier runs.
func TestBoard_WorkflowContext(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.BaseURL = "https://task.example.com/"
	f.board.AgentCommand = []string{"sh", "-c", `cat "$TASK_UI_PROMPT_WORKFLOW" > "$TASK_UI_PROMPT_REPLY"`}
	wf := &model.Workflow{Name: "watering", Prompt: "water today?"}
	flowID := model.NewID()

	newRun := func(input string) *model.WorkflowRun {
		run := &model.WorkflowRun{WorkflowID: flowID, Name: wf.Name, UserID: admin.ID, UserName: admin.Name, Input: input}
		assert.NoError(t, f.board.WorkflowRuns.Create(ctx, run, nil))
		return run
	}
	asked := func(run *model.WorkflowRun) string {
		answer, err := askFinal(f.board, ctx, admin, run, wf, run.Input)
		assert.NoError(t, err)
		return answer
	}
	first := newRun("porec")
	text := asked(first)
	assert.True(t, strings.HasPrefix(text, "## Workflow context\n\nThe workflow watering, https://task.example.com/workflows/"+flowID+".\n"), text)
	assert.Contains(t, text, "This is its run 1, https://task.example.com/workflows/"+flowID+"/runs/"+first.ID+", started ")
	assert.Contains(t, text, " by admin, with the input:\n\nporec\n")
	assert.Contains(t, text, "There are no earlier runs with an answer to compare with.")

	// Five earlier runs: one still running, one without an answer; the
	// block keeps the three newest answers.
	stamp := time.Date(2026, 10, 1, 6, 0, 0, 0, time.UTC)
	for i, answer := range []string{"day one", "day two", "", "day four", strings.Repeat("x", 2000) + "TAIL"} {
		run := newRun("")
		run.State, run.Answer = model.StateOK, answer
		run.SetStartedAt(stamp.Add(time.Duration(i) * 24 * time.Hour))
		assert.NoError(t, f.board.WorkflowRuns.Update(ctx, run))
	}
	running := newRun("")
	running.Answer = "still going"
	assert.NoError(t, f.board.WorkflowRuns.Update(ctx, running))

	now := newRun("")
	text = asked(now)
	assert.Contains(t, text, "This is its run 8,")
	assert.Contains(t, text, ", with no input.\n")
	assert.Contains(t, text, "The answers of its last 3 finished runs, newest first")
	five := strings.Index(text, "### The run of 2026-10-05T06:00:00Z, ok")
	four := strings.Index(text, "### The run of 2026-10-04T06:00:00Z, ok\n\n```\nday four\n```")
	two := strings.Index(text, "### The run of 2026-10-02T06:00:00Z, ok\n\n```\nday two\n```")
	assert.True(t, five >= 0 && five < four && four < two, text)
	assert.NotContains(t, text, "day one")
	assert.NotContains(t, text, "still going")
	assert.NotContains(t, text, "TAIL")
	assert.Contains(t, text, "The answer is cut to its first 1.5 KiB.")
	assert.Contains(t, text, "\nWorkflow run "+now.ID+" of the workflow watering")
	assert.Equal(t, 3, note.EarlierAnswers)
}
