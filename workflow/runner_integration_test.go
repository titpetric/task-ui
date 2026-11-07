//go:build integration

package workflow_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// The run of a workflow's step sits under the workflow run, so the
// sidebar's roots leave it out and the workflow run's page lists it.
func TestRunnerStepParent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	defer db.Close()

	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{Runs: runs, Recordings: &recording.Store{Root: t.TempDir()}})
	infra := &folder.Folder{ID: "infra", Dir: t.TempDir(), Jobs: []*folder.Job{job("build", "echo built")}}
	workflows, flows := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	wf := &model.Workflow{Name: "release"}
	assert.NoError(t, workflows.Create(ctx, wf, []model.WorkflowStep{{Kind: model.StepJob, Folder: "infra", Job: "build"}}))

	r := workflow.New(workflows, flows, catalogue{"infra": infra}, manager)
	run, err := r.Start(ctx, wf, &model.User{ID: "u1", Name: "ops"}, "")
	assert.NoError(t, err)
	<-r.Done(run.ID)

	children, err := runs.Children(ctx, run.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, "build", children[0].Job)
	assert.Equal(t, storage.ParentWorkflowRun, children[0].ParentKind)

	roots, err := runs.List(ctx, storage.RunFilter{Roots: true})
	assert.NoError(t, err)
	assert.Len(t, roots, 0, "the sidebar lists the workflow run and not its step")
}
