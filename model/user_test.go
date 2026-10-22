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

	// The internal folder is the admins', whatever the folder list says.
	admin := &model.User{Name: "root", Role: model.RoleAdmin, Folders: `["web"]`}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.True(t, admin.CanSee(model.Internal))
	assert.True(t, admin.CanSee("internal/tools"))
	assert.False(t, ops.CanSee(model.Internal))
	assert.False(t, guest.CanSee("internal/tools"))
	assert.False(t, (&model.User{Role: model.RoleRunner, Folders: `["internal"]`}).CanSee(model.Internal))
}

func TestUser_Listed(t *testing.T) {
	t.Parallel()
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	locked := []string{"ops"}
	admins := []string{"admin"}

	assert.True(t, ops.Listed(nil, nil))
	assert.True(t, ops.Listed(locked, nil))
	assert.False(t, ops.Listed(locked, admins))
	assert.False(t, (&model.User{Name: "dev", Role: model.RoleRunner}).Listed(locked, nil))
	assert.True(t, ops.Listed([]string{"runner"}, nil), "a role in the allow list")
	assert.False(t, (*model.User)(nil).Listed(nil, nil))
}

func TestUser_CanAdmin(t *testing.T) {
	t.Parallel()
	assert.True(t, (&model.User{Role: model.RoleAdmin}).CanAdmin())
	assert.False(t, (&model.User{Role: model.RoleRunner}).CanAdmin())
	assert.False(t, (*model.User)(nil).CanAdmin())
}

func TestUser_SeesRun(t *testing.T) {
	t.Parallel()
	limited := &model.User{ID: "u1", Name: "rob", Folders: `["malta"]`}
	other := &model.User{ID: "u2", Name: "consuela"}
	build := &model.Run{Folder: "malta", Job: "build", UserID: "u2"}
	prompt := &model.Run{Folder: "malta", Job: model.PromptJob, UserID: "u2"}
	mine := &model.Run{Folder: "infra", Job: model.PromptJob, UserID: "u1"}
	assert.True(t, limited.SeesRun(build, false))
	assert.False(t, limited.SeesRun(prompt, false), "a prompt's run takes prompt/view")
	assert.True(t, limited.SeesRun(prompt, true))
	assert.True(t, limited.SeesRun(mine, false), "the asker sees their own")
	assert.False(t, other.SeesRun(mine, false))
	assert.True(t, other.SeesRun(mine, true))
	assert.False(t, limited.SeesRun(&model.Run{Folder: "infra", Job: "build"}, true))
	assert.False(t, (*model.User)(nil).SeesRun(build, true))
}
