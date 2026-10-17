package app_test

import (
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUserEntry_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, app.UserEntry{Name: "ops", Password: "x", Role: model.RoleRunner}.Validate())
	assert.NoError(t, app.UserEntry{Name: "ops", Password: "x"}.Validate())
	assert.ErrorContains(t, app.UserEntry{Password: "x"}.Validate(), "needs a name")
	assert.ErrorContains(t, app.UserEntry{Name: "ops"}.Validate(), "password is needed")
	assert.ErrorContains(t, app.UserEntry{Name: "ops", Password: "x", Role: "king"}.Validate(), `role "king"`)
}

func TestUserEntry_Hash(t *testing.T) {
	t.Parallel()

	hashed, err := access.HashPassword("demo")
	assert.NoError(t, err)
	same, err := app.UserEntry{Password: hashed}.Hash()
	assert.NoError(t, err)
	assert.Equal(t, hashed, same)

	fresh, err := app.UserEntry{Password: "plain"}.Hash()
	assert.NoError(t, err)
	assert.True(t, access.Hashed(fresh))
	assert.True(t, access.CheckPassword(&model.User{Password: fresh}, "plain"))
}

func TestUserEntry_User(t *testing.T) {
	t.Parallel()

	user, err := app.UserEntry{Name: " ops ", Password: "plain", Folders: []string{"infra"}}.User()
	assert.NoError(t, err)
	assert.Equal(t, "ops", user.Name)
	assert.Equal(t, model.RoleViewer, user.Role)
	assert.Equal(t, []string{"infra"}, user.FolderList())
	assert.True(t, access.CheckPassword(user, "plain"))
}
