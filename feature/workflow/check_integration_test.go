//go:build integration

package workflow_test

import (
	"context"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// A check step reads the step it follows and sends the steps round again
// while the marker is there, as many times as its tries allow: two
// passes of the prompt before it, then the step after it, once. The
// check's own answer says what it found, and the run ends ok.
func TestRunnerCheckRerun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	defer db.Close()

	workflows, runs := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	wf := &model.Workflow{Name: "refine"}
	assert.NoError(t, workflows.Create(ctx, wf, []model.WorkflowStep{
		{Kind: model.StepPrompt, Folder: model.Default, Job: model.PromptJob, Prompt: "refine the design"},
		{Kind: model.StepCheck, Markers: "Next:", Tries: 2},
		{Kind: model.StepPrompt, Folder: model.Default, Job: model.PromptJob, Prompt: "write it up"},
	}))

	c := &claude{text: "the first half is done.\nNext: the second half\n"}
	r := workflow.New(workflows, runs, nil, nil)
	r.Prompter, r.Output = c, c

	run, err := r.Start(ctx, wf, &model.User{ID: "u1", Name: "ops"}, "go")
	assert.NoError(t, err)
	<-r.Done(run.ID)

	ended, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, ended.State)
	assert.Equal(t, 3, c.count(), "the prompt ran twice, and the step after the check once")

	rows, err := runs.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(rows[1].Answer, `found "Next:" in step 1`), rows[1].Answer)
	assert.True(t, strings.Contains(rows[1].Answer, "pass 2 of 2 was the last"), rows[1].Answer)
	assert.Equal(t, model.StateOK, rows[2].State)
}

// Two checks of one project each spend the tries they declare: the second
// one's first look is its own pass 1, not the pass the first one left.
func TestRunnerCheckTwice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	defer db.Close()

	workflows, runs := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	wf := &model.Workflow{Name: "twice"}
	assert.NoError(t, workflows.Create(ctx, wf, []model.WorkflowStep{
		{Kind: model.StepPrompt, Folder: model.Default, Job: model.PromptJob, Prompt: "refine the design"},
		{Kind: model.StepCheck, Markers: "Next:", Tries: 2},
		{Kind: model.StepPrompt, Folder: model.Default, Job: model.PromptJob, Prompt: "write it up"},
		{Kind: model.StepCheck, Markers: "Next:", Tries: 2},
	}))

	c := &claude{text: "the first half is done.\nNext: the second half\n"}
	r := workflow.New(workflows, runs, nil, nil)
	r.Prompter, r.Output = c, c

	run, err := r.Start(ctx, wf, &model.User{ID: "u1", Name: "ops"}, "go")
	assert.NoError(t, err)
	<-r.Done(run.ID)

	rows, err := runs.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(rows[3].Answer, "pass 2 of 2 was the last"), rows[3].Answer)
	// Three of the first prompt and two of the write-up: the second check
	// sends the project round on its own first look, which it would not
	// have left to spend had it read the first check's count.
	assert.Equal(t, 5, c.count(), "each check spent its own two passes")
	ended, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, ended.State)
}
