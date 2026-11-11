package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/workflow"
)

// users holds users by ID, each with the permissions of their role.
type users map[string]*model.User

func (u users) Get(_ context.Context, id string) (*model.User, error) {
	if user, ok := u[id]; ok {
		return user, nil
	}
	return nil, errors.New("no user")
}

func (u users) Permissions(_ context.Context, user *model.User) permission.Set {
	if user.Role == model.RoleAdmin {
		return permission.Set{User: user}
	}
	return permission.Set{User: user, Role: permission.Policy{"run": permission.Allow}}
}

// A flow says who its runs run as, the user set on it or the starter,
// and which step the starter or that user may not run.
func TestFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	infra := &folder.Folder{ID: "infra", Jobs: []*folder.Job{job("plan", "echo planned"), job("deploy", "echo deployed")}}
	infra.Job("deploy").Allow = []string{model.RoleAdmin}
	r := workflow.New(nil, nil, catalogue{"infra": infra}, nil)
	admin := &model.User{ID: "a", Name: "admin", Role: model.RoleAdmin}
	ops := &model.User{ID: "o", Name: "ops", Role: model.RoleRunner}
	steps := []model.WorkflowStep{{Folder: "infra", Job: "deploy"}}

	open := r.Flow(&model.Workflow{Name: "open", UserName: "admin"}, steps)
	assert.Equal(t, "open", open.Workflow().Name)
	assert.Len(t, open.Steps(), 1)
	runAs, err := open.RunAs(ctx, ops)
	assert.NoError(t, err)
	assert.Equal(t, "ops", runAs.Name, "the starter")
	assert.Equal(t, "", open.RunAsName(ctx))
	assert.Equal(t, "infra / deploy", open.MayRun(ctx, users{}.Permissions(ctx, ops)))
	_, refusal, err := open.Check(ctx, ops, users{}.Permissions(ctx, ops))
	assert.NoError(t, err)
	assert.Equal(t, workflow.Refusal{Step: "infra / deploy"}, refusal)

	bare := r.Flow(&model.Workflow{Name: "bare", UserName: "admin", RunAsUserID: "o"}, steps)
	runAs, err = bare.RunAs(ctx, admin)
	assert.NoError(t, err)
	assert.Equal(t, "o", runAs.ID, "without Users, the set user by ID")

	r.Users = users{"a": admin, "o": ops}
	asOps := r.Flow(&model.Workflow{Name: "as ops", RunAsUserID: "o"}, steps)
	assert.Equal(t, "ops", asOps.RunAsName(ctx))
	runAs, refusal, err = asOps.Check(ctx, admin, users{}.Permissions(ctx, admin))
	assert.NoError(t, err)
	assert.Equal(t, "ops", runAs.Name)
	assert.Equal(t, workflow.Refusal{Step: "infra / deploy", User: "ops"}, refusal)

	asAdmin := r.Flow(&model.Workflow{Name: "as admin", RunAsUserID: "a"}, []model.WorkflowStep{{Folder: "infra", Job: "plan"}})
	runAs, refusal, err = asAdmin.Check(ctx, ops, users{}.Permissions(ctx, ops))
	assert.NoError(t, err)
	assert.Equal(t, "admin", runAs.Name)
	assert.Equal(t, workflow.Refusal{}, refusal)

	gone := r.Flow(&model.Workflow{Name: "gone", RunAsUserID: "nobody"}, steps)
	assert.Equal(t, "", gone.RunAsName(ctx))
	_, _, err = gone.Check(ctx, admin, users{}.Permissions(ctx, admin))
	assert.ErrorIs(t, err, workflow.ErrRunAsGone)
}
