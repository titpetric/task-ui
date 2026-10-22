package permission_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

func TestSet_Allowed(t *testing.T) {
	t.Parallel()
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	role := permission.Policy{"note/all": permission.Allow, "run": permission.Allow}

	set := permission.Set{User: ops, Role: role}
	assert.True(t, set.Allowed("note/edit"))
	assert.True(t, set.Allowed("run"))
	assert.False(t, set.Allowed("secret/view"), "what nobody sets is denied")

	set.Own = permission.Policy{"note/delete": permission.Deny, "secret/view": permission.Allow}
	assert.False(t, set.Allowed("note/delete"), "the user's deny beats the role's allow")
	assert.True(t, set.Allowed("note/edit"))
	assert.True(t, set.Allowed("secret/view"), "the user's allow adds to the role")

	set.Role = permission.Policy{"secret/all": permission.Deny}
	assert.False(t, set.Allowed("secret/view"), "the role's deny beats the user's allow")

	admin := permission.Set{User: &model.User{Name: "root", Role: model.RoleAdmin}, Own: permission.Policy{"run": permission.Deny}}
	assert.True(t, admin.Allowed("run"), "an admin may do everything")
	assert.True(t, admin.Allowed("secret/delete"))
	assert.False(t, permission.Set{}.Allowed("note/view"), "nobody may do nothing")
}

func TestSet_Run(t *testing.T) {
	t.Parallel()
	runner := permission.Policy{"run": permission.Allow}
	ops := permission.Set{User: &model.User{Name: "ops", Role: model.RoleRunner}, Role: runner}
	assert.True(t, ops.Run("infra", nil, nil))
	assert.True(t, ops.Run("infra", []string{"ops"}, []string{"runner"}))
	assert.False(t, ops.Run("infra", []string{"dev"}, nil))
	assert.False(t, ops.Run(model.Internal, nil, nil), "the internal folder is the admins'")
	assert.False(t, permission.Set{User: &model.User{Name: "guest", Role: model.RoleViewer}}.Run("infra", nil, nil))

	admin := permission.Set{User: &model.User{Name: "root", Role: model.RoleAdmin}}
	assert.True(t, admin.Run(model.Internal, []string{"ops"}, []string{"dev"}))
	assert.False(t, permission.Set{}.Run("infra", nil, nil))
}

func TestSet_Effective(t *testing.T) {
	t.Parallel()
	set := permission.Set{User: &model.User{Name: "ops", Role: model.RoleRunner}, Role: permission.Policy{"note/all": permission.Allow}, Own: permission.Policy{"note/delete": permission.Deny}}
	effective := set.Effective()
	assert.True(t, effective["note/view"])
	assert.False(t, effective["note/delete"])
	assert.False(t, effective["run"])
	_, ok := effective["note/all"]
	assert.False(t, ok, "no check names all")
	assert.True(t, permission.Set{User: &model.User{Role: model.RoleAdmin}}.Effective()["shell"])
}
