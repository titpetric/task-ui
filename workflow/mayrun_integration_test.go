//go:build integration

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

// MayRun names the first step a user may not run: a job the allow
// lists keep from them, a prompt without prompt/edit, a host job
// without hosts, a job nothing holds, the steps of a nested workflow
// too; an admin runs them all.
func TestRunner_MayRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	inner := f.create(t, "build")
	infra := &folder.Folder{ID: "infra", Jobs: []*folder.Job{job("build", "echo built"), job("test", "echo tested")}}
	r := workflow.New(f.workflows, f.wfRuns, catalogue{"infra": infra}, nil)
	steps := []model.WorkflowStep{{Folder: "infra", Job: "test"}, {WorkflowRef: inner.ID}, {Kind: model.StepPrompt, Folder: "infra", Prompt: "why"}}

	runner := permission.Set{User: &model.User{ID: "u1", Name: "ops", Role: model.RoleRunner}, Role: permission.Policy{"run": permission.Allow, "prompt/edit": permission.Allow}}
	assert.Equal(t, "", r.MayRun(ctx, runner, steps))

	noPrompt := permission.Set{User: runner.User, Role: permission.Policy{"run": permission.Allow}}
	assert.Equal(t, "the prompt in infra", r.MayRun(ctx, noPrompt, steps))

	viewer := permission.Set{User: &model.User{ID: "u2", Name: "guest", Role: model.RoleViewer}, Role: permission.Policy{"prompt/edit": permission.Allow}}
	assert.Equal(t, "infra / test", r.MayRun(ctx, viewer, steps))

	infra.Job("build").Allow = []string{"admin"}
	assert.Equal(t, "infra / build", r.MayRun(ctx, runner, steps), "the nested workflow's step")

	admin := permission.Set{User: &model.User{ID: "u3", Name: "root", Role: model.RoleAdmin}}
	assert.Equal(t, "", r.MayRun(ctx, admin, steps))

	assert.Equal(t, "infra / gone", r.MayRun(ctx, admin, []model.WorkflowStep{{Folder: "infra", Job: "gone"}}))
	assert.Equal(t, "infra / "+folder.ShellID, r.MayRun(ctx, admin, []model.WorkflowStep{{Folder: "infra", Job: folder.ShellID}}))
	hosts := []model.WorkflowStep{{Folder: model.Default, Job: "@host:*"}}
	assert.Equal(t, "_ / @host:*", r.MayRun(ctx, runner, hosts), "no root folder in the catalogue")
}
