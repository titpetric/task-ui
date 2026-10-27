package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// envBlock returns the Environment block of a folder page, "" without one.
func envBlock(body string) string {
	at := strings.Index(body, "<h3>Environment")
	if at < 0 {
		return ""
	}
	rest := body[at:]
	return rest[:strings.Index(rest, "</section>")]
}

// A project with no env file still draws the Environment block, with the
// button that adds one; with files, the table lists them.
func TestFolder_EnvironmentWhileEmpty(t *testing.T) {
	t.Parallel()

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	block := envBlock(render(t, "folder", p))
	assert.Contains(t, block, `<p class="empty-line">No env files.</p>`)
	assert.Contains(t, block, `href="/f/internal/apps/env"`)
	assert.NotContains(t, block, "<table")

	p.EnvFiles = []envfile.File{{Folder: "apps", Path: ".env"}}
	block = envBlock(render(t, "folder", p))
	assert.NotContains(t, block, "No env files.")
	assert.Contains(t, block, "<table")
	assert.Contains(t, block, `href="/f/internal/apps/env"`)
}

// The empty block is a project's alone; a workspace or a folder inside a
// project draws it only for files it has, and a user without env/edit
// sees none.
func TestFolder_EnvironmentRules(t *testing.T) {
	t.Parallel()

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	p.Folder = &folder.Folder{ID: "apps/api", Title: "api"}
	p.FolderMenu = &view.FolderMenu{ID: "apps/api"}
	assert.Equal(t, "", envBlock(render(t, "folder", p)), "a plain folder has no empty block")

	p.Folder = &folder.Folder{ID: "malta", Title: "Malta"}
	p.FolderMenu = &view.FolderMenu{ID: "malta", Workspace: true}
	assert.Equal(t, "", envBlock(render(t, "folder", p)), "env files are per project")
	assert.NotContains(t, pageHead(t, render(t, "folder", p)), "Add env file")

	p = projectPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	assert.Equal(t, "", envBlock(render(t, "folder", p)), "adding an env file takes env/edit")

	p.EnvFiles = []envfile.File{{Folder: "apps", Path: ".env"}}
	block := envBlock(render(t, "folder", p))
	assert.Contains(t, block, "<table", "the files list for a viewer")
	assert.NotContains(t, block, "Add env file")
}

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
