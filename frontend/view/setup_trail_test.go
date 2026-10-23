package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// The footer links the steps either side among those the user opens,
// and leaves out a side with none.
func TestSetupTrail(t *testing.T) {
	t.Parallel()

	p := &view.Page{}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)
	assert.Nil(t, p.SetupTrail())

	p.Step = view.StepIntegrations
	trail := p.SetupTrail()
	assert.Equal(t, "Secrets", trail.Previous.Name)
	assert.Equal(t, "Users", trail.Next.Name)
	assert.Equal(t, 10, len(trail.Steps))

	p.Step = view.StepInstance
	trail = p.SetupTrail()
	assert.Nil(t, trail.Previous)
	assert.Equal(t, "/settings/retention", trail.Next.Href)

	p.Step = view.StepProjects
	trail = p.SetupTrail()
	assert.Equal(t, "Workspaces", trail.Previous.Name)
	assert.Nil(t, trail.Next)

	// A limited user skips the steps they may not open.
	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow, "workflow/view": permission.Allow})
	p.Step = view.StepSecrets
	trail = p.SetupTrail()
	assert.Equal(t, "Data retention", trail.Previous.Name)
	assert.Equal(t, "Workflows", trail.Next.Name)

	// No footer on a step the user may not open, or with no other step.
	p.Step = view.StepKeys
	assert.Nil(t, p.SetupTrail())
	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, permission.Policy{"workflow/view": permission.Allow})
	p.Step = view.StepWorkflows
	assert.Nil(t, p.SetupTrail())
}
