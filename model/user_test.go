package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUser_Initials(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"Admin Back":  "AB",
		"admin":       "AD",
		"tit.petric":  "TP",
		"ops@example": "OE",
		"x":           "X",
		"":            "?",
		"...":         "?",
	} {
		assert.Equal(t, want, (&model.User{Name: name}).Initials(), name)
	}
	assert.Equal(t, "?", (*model.User)(nil).Initials())
}

func TestUser_FolderList(t *testing.T) {
	t.Parallel()
	assert.Len(t, (*model.User)(nil).FolderList(), 0)
	assert.Equal(t, []string{"web"}, (&model.User{Folders: `["web"]`}).FolderList())
}

func TestUser_CanSee(t *testing.T) {
	t.Parallel()
	scoped := &model.User{Name: "web", Role: model.RoleRunner, Folders: `["web"]`}
	guest := &model.User{Name: "guest", Role: model.RoleViewer}
	assert.False(t, scoped.CanSee("infra"))
	assert.True(t, scoped.CanSee("web"))
	assert.True(t, guest.CanSee("infra"))
	assert.False(t, (*model.User)(nil).CanSee("web"))
}

func TestUser_CanRun(t *testing.T) {
	t.Parallel()
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	dev := &model.User{Name: "dev", Role: model.RoleRunner}
	scoped := &model.User{Name: "web", Role: model.RoleRunner, Folders: `["web"]`}
	guest := &model.User{Name: "guest", Role: model.RoleViewer}
	locked := []string{"ops"}
	admins := []string{"admin"}

	assert.True(t, admin.CanRun("infra", locked, admins))
	assert.True(t, ops.CanRun("infra", locked, nil))
	assert.False(t, ops.CanRun("infra", locked, admins))
	assert.True(t, dev.CanRun("web", nil, nil))
	assert.False(t, dev.CanRun("infra", locked, nil))
	assert.True(t, scoped.CanRun("web", nil, nil))
	assert.False(t, scoped.CanRun("infra", nil, nil))
	assert.False(t, guest.CanRun("web", nil, nil))
	assert.False(t, (*model.User)(nil).CanRun("web", nil, nil))
	assert.True(t, ops.CanRun("infra", []string{"runner"}, nil), "a role in the allow list")
}

func TestUser_CanAdmin(t *testing.T) {
	t.Parallel()
	assert.True(t, (&model.User{Role: model.RoleAdmin}).CanAdmin())
	assert.False(t, (&model.User{Role: model.RoleRunner}).CanAdmin())
	assert.False(t, (*model.User)(nil).CanAdmin())
}
