package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// The folder's menu schedules a run of it, for a project and a
// workspace with schedule/edit.
func TestFolderMenu_AddSchedule(t *testing.T) {
	t.Parallel()

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	head := pageHead(t, render(t, "folder", p))
	assert.Contains(t, head, `href="/schedules?folder=apps"`)
	assert.Contains(t, head, "<span>Add schedule</span>")

	p.Folder = &folder.Folder{ID: "malta", Title: "Malta"}
	p.FolderMenu = &view.FolderMenu{ID: "malta", Workspace: true}
	assert.Contains(t, pageHead(t, render(t, "folder", p)), `href="/schedules?folder=malta"`)

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	assert.NotContains(t, pageHead(t, render(t, "folder", p)), "/schedules", "scheduling takes schedule/edit")
}
