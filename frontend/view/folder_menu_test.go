package view_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// The menu links the settings of a workspace or of a project.
func TestFolderMenu_Settings(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/settings/workspaces/malta", (&view.FolderMenu{ID: "malta", Workspace: true}).Settings())
	assert.Equal(t, "/settings/projects/malta/api", (&view.FolderMenu{ID: "malta/api", Project: true}).Settings())
	assert.Equal(t, "", (&view.FolderMenu{ID: "apps/web"}).Settings())
	assert.Equal(t, "", (*view.FolderMenu)(nil).Settings())
}

// Remove checkout names the folder and where to clone it again from.
func TestRemoveConfirm(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Remove the folder apps and its checkout? You can clone it again from git@example.com:org/apps.git.", view.RemoveConfirm("apps", "git@example.com:org/apps.git"))
	assert.Contains(t, view.RemoveConfirm("apps", ""), "has no remote")
}

// pageHead returns the page head of a rendered page.
func pageHead(t *testing.T, body string) string {
	t.Helper()
	start := strings.Index(body, `<div class="page-head">`)
	assert.True(t, start >= 0, "a page head")
	end := strings.Index(body[start:], "\n</div>")
	return body[start : start+end]
}

// projectPage is the page of the project apps, a checkout with a pull
// job of the git skill, as user sees it with the role's policy.
func projectPage(t *testing.T, user *model.User, role permission.Policy) *view.Page {
	t.Helper()
	p := admin(t)
	pull := &folder.Job{Folder: "apps", ID: "@git:pull"}
	pull.Name = "pull"
	p.Folder = &folder.Folder{ID: "apps", Title: "apps"}
	signIn(p, user, role)
	p.Checkout = &git.Info{Remote: "git@example.com:org/apps.git", Branch: "main"}
	p.FolderMenu = &view.FolderMenu{ID: "apps", Project: true, Pull: pull}
	return p
}

// A project's page has the [...] menu in its head for a user with the
// project rules: settings, Pull now, and Remove checkout last.
func TestFolderMenu_Project(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "folder", projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)))
	head := pageHead(t, buf.String())
	settings := strings.Index(head, `href="/settings/projects/apps"`)
	pull := strings.Index(head, `data-job="@git:pull"`)
	remove := strings.Index(head, `data-delete="/api/projects/apps"`)
	assert.True(t, settings > 0 && pull > settings && remove > pull, "settings, Pull now, then Remove checkout")
	assert.Contains(t, head, `<span>Remove checkout</span>`)
	assert.Contains(t, head, `class="more-item danger"`)
	assert.Contains(t, head, `data-confirm="Remove the folder apps and its checkout? You can clone it again from git@example.com:org/apps.git."`)
	assert.Contains(t, head, `data-next="/settings/projects?notice=Removed+apps"`)

	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", projectPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})))
	head = pageHead(t, buf.String())
	assert.NotContains(t, head, "/settings/projects/apps")
	assert.NotContains(t, head, "data-delete")
	assert.Contains(t, head, `data-job="@git:pull"`, "a runner may still pull")

	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", projectPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"project/view": permission.Allow, "project/delete": permission.Allow})))
	head = pageHead(t, buf.String())
	assert.Contains(t, head, `href="/settings/projects/apps"`, "the rules, not the role, decide")
	assert.Contains(t, head, `data-delete="/api/projects/apps"`)
	assert.NotContains(t, head, "@git:pull", "without run, no Pull now")

	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", projectPage(t, &model.User{Name: "guest", Role: model.RoleViewer}, nil)))
	assert.NotContains(t, pageHead(t, buf.String()), "more-menu", "nothing to show, no menu")
}

// A workspace's page links its settings, with nothing to remove.
func TestFolderMenu_Workspace(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	p.Folder = &folder.Folder{ID: "malta", Title: "Malta"}
	p.FolderMenu = &view.FolderMenu{ID: "malta", Workspace: true}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "folder", p))
	head := pageHead(t, buf.String())
	assert.Contains(t, head, `href="/settings/workspaces/malta"`)
	assert.Contains(t, head, "Workspace settings")
	assert.NotContains(t, head, "data-delete")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", p))
	assert.NotContains(t, pageHead(t, buf.String()), "more-menu", "a workspace has no Pull now")
}

// The configure page of a checkout has Remove checkout in its head, for
// a user with project/delete.
func TestFolderMenu_ProjectEdit(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "project_edit", p))
	head := pageHead(t, buf.String())
	assert.Contains(t, head, `data-delete="/api/projects/apps"`)
	assert.Contains(t, head, `data-confirm="Remove the folder apps and its checkout? You can clone it again from git@example.com:org/apps.git."`)
	assert.Contains(t, head, `data-next="/settings/projects?notice=Removed+apps"`)

	p.Checkout = nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "project_edit", p))
	assert.NotContains(t, pageHead(t, buf.String()), "data-delete", "a local folder has no checkout to remove")
}

// Delete of a workspace names the projects that go with it and where
// to clone it again from.
func TestDeleteWorkspaceConfirm(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Delete the workspace acme with its checkout, keys and secrets, and the projects inside it: acme/api, acme/web? You can clone it again from git@example.com:acme.git.", view.DeleteWorkspaceConfirm("acme", "git@example.com:acme.git", []string{"acme/api", "acme/web"}))
	assert.Equal(t, "Delete the workspace labs with its checkout, keys and secrets? You can clone it again from file:///labs.git.", view.DeleteWorkspaceConfirm("labs", "file:///labs.git", nil))
}
