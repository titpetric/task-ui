package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// stepNames returns the names of steps, in order.
func stepNames(steps []view.SetupStep) []string {
	names := []string{}
	for _, step := range steps {
		names = append(names, step.Name)
	}
	return names
}

// An admin opens every step, in setup order; a limited user the steps
// their permissions allow, numbered among themselves.
func TestSetupStep(t *testing.T) {
	t.Parallel()

	p := &view.Page{Step: view.StepUsers}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)
	steps := p.SetupSteps()
	assert.Equal(t, []string{"Instance", "Keys", "Secrets", "Integrations", "Users", "Permissions", "Workflows", "Workspaces", "Projects"}, stepNames(steps))
	assert.Equal(t, view.SetupStep{Name: "Users", Href: "/settings/users", Icon: "users", Desc: steps[4].Desc, Number: 5, Current: true}, steps[4])
	assert.False(t, steps[0].Current)

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow, "workflow/view": permission.Allow})
	steps = p.SetupSteps()
	assert.Equal(t, []string{"Instance", "Secrets", "Workflows"}, stepNames(steps))
	assert.Equal(t, 3, steps[2].Number)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, permission.Policy{"workflow/view": permission.Allow})
	assert.Equal(t, []string{"Workflows"}, stepNames(p.SetupSteps()))
}
