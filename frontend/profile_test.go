package frontend

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// A profile and its edit page are where a user's name says they are.
func TestProfileHref(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/users/ops", profileHref("ops"))
	assert.Equal(t, "/users/ops/edit", editHref("ops"))
	assert.Equal(t, "/users/a%20b/edit", editHref("a b"))
}

// An admin edits by role, and so does a reader whose role allows
// user/edit. A rule of the reader's own alone does not: granting yourself
// more would be the next thing it allowed.
func TestEditsByRole(t *testing.T) {
	t.Parallel()
	admin := &view.Page{User: &model.User{Name: "root", Admin: true}}
	assert.True(t, editsByRole(admin))

	byRole := &view.Page{
		User:        &model.User{Name: "ops"},
		Permissions: permission.Set{Role: permission.Policy{"user/edit": permission.Allow}},
	}
	assert.True(t, editsByRole(byRole))

	byOwn := &view.Page{
		User:        &model.User{Name: "ops"},
		Permissions: permission.Set{Own: permission.Policy{"user/edit": permission.Allow}},
	}
	assert.False(t, editsByRole(byOwn))
}
