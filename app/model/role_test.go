package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRoles(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []model.Role{model.RoleAdmin, model.RoleRunner, model.RoleViewer}, model.Roles())
}

func TestValidRole(t *testing.T) {
	t.Parallel()
	for _, role := range model.Roles() {
		assert.True(t, model.ValidRole(role), role)
	}
	assert.False(t, model.ValidRole("root"))
	assert.False(t, model.ValidRole(""))
}
