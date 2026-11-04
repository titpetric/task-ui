package workflow_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/workflow"
)

// A check step and a disabled one take no permission, and a check names
// no folder, so the job lookup a job step gets must not reach it. The
// first step the permissions refuse is the answer, by its folder and job.
func TestRunner_MayRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	infra := &folder.Folder{ID: "infra", Dir: t.TempDir(), Jobs: []*folder.Job{job("plan", "echo planned")}}
	r := workflow.New(nil, nil, catalogue{"infra": infra}, nil)

	quiet := []model.WorkflowStep{
		{Kind: model.StepCheck, Markers: "Next:"},
		{Folder: "infra", Job: "plan", Disabled: true},
	}
	assert.Equal(t, "", r.MayRun(ctx, permission.Set{}, &model.Workflow{}, quiet))

	asks := append(quiet, model.WorkflowStep{Folder: "infra", Job: "plan"})
	assert.Equal(t, "infra / plan", r.MayRun(ctx, permission.Set{}, &model.Workflow{}, asks))
}
