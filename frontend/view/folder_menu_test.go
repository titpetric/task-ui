package view_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

// A workspace's menu links its SSH keys and its secrets; a project's
// has neither.
func TestFolderMenu_ScopedPages(t *testing.T) {
	t.Parallel()

	ws := &view.FolderMenu{ID: "malta", Workspace: true}
	assert.Equal(t, "/settings/workspaces/malta/keys", ws.KeysPage())
	assert.Equal(t, "/settings/secrets?workspace=malta", ws.SecretsPage())
	assert.Equal(t, "/settings/workspaces/internal/keys", (&view.FolderMenu{ID: "internal", Workspace: true}).KeysPage())
	project := &view.FolderMenu{ID: "malta/api", Project: true}
	assert.Equal(t, "", project.KeysPage())
	assert.Equal(t, "", project.SecretsPage())
	assert.Equal(t, "", (*view.FolderMenu)(nil).SecretsPage())
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
	assert.Contains(t, head, `href="/settings/projects/apps"`, "the rules decide")
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
	assert.Contains(t, head, "Edit workspace")
	assert.Contains(t, head, `href="/settings/workspaces/malta/keys"`)
	assert.Contains(t, head, `href="/settings/secrets?workspace=malta"`)
	assert.NotContains(t, head, "data-delete")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow})
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", p))
	head = pageHead(t, buf.String())
	assert.Contains(t, head, `href="/settings/secrets?workspace=malta"`, "secret/view links the secrets")
	assert.NotContains(t, head, `href="/settings/workspaces/malta/keys"`, "the keys take workspace/view")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", p))
	head = pageHead(t, buf.String())
	assert.Contains(t, head, `href="/settings/workspaces/malta/keys"`)
	assert.NotContains(t, head, `href="/settings/secrets?workspace=malta"`)

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "folder", p))
	assert.NotContains(t, pageHead(t, buf.String()), "more-menu", "a workspace has no Pull now")
}

// A workspace's page adds a project to itself: in the menu, and as the
// main button of the head while it has none.
func TestFolderMenu_WorkspaceAddProject(t *testing.T) {
	t.Parallel()

	p := projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	p.Folder = &folder.Folder{ID: "malta", Title: "Malta"}
	p.FolderMenu = &view.FolderMenu{ID: "malta", Workspace: true}
	head := pageHead(t, render(t, "folder", p))
	assert.Contains(t, head, `<a class="btn btn-primary" href="/settings/projects/new?workspace=malta">`)
	assert.Contains(t, head, `<a class="more-item" role="menuitem" href="/settings/projects/new?workspace=malta"`)

	p.FolderMenu.Projects = 2
	head = pageHead(t, render(t, "folder", p))
	assert.NotContains(t, head, `class="btn btn-primary"`, "with projects, the menu alone adds one")
	assert.Contains(t, head, `href="/settings/projects/new?workspace=malta"`)

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	p.FolderMenu.Projects = 0
	head = pageHead(t, render(t, "folder", p))
	assert.NotContains(t, head, "/settings/projects/new", "adding a project takes project/edit")

	p.FolderMenu = &view.FolderMenu{ID: "malta/api", Project: true}
	signIn(p, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	assert.NotContains(t, pageHead(t, render(t, "folder", p)), "/settings/projects/new", "a project adds no project")
}

func TestFolderMenu_AddProjectPage(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/settings/projects/new?workspace=malta", (&view.FolderMenu{ID: "malta", Workspace: true}).AddProjectPage())
	assert.Equal(t, "", (&view.FolderMenu{ID: "malta/api", Project: true}).AddProjectPage())
	assert.Equal(t, "", (*view.FolderMenu)(nil).AddProjectPage())
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

// Delete of a workspace names the projects that go with it, their
// checkouts too.
func TestDeleteBlocked(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Delete its 1 variable first", view.DeleteBlocked(1))
	assert.Equal(t, "Delete its 3 variables first", view.DeleteBlocked(3))
}

func TestDeleteWorkspaceConfirm(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Delete the workspace acme with its keys and secrets, and the projects inside it with their checkouts: acme/api, acme/web? Its directory goes once nothing else is in it.", view.DeleteWorkspaceConfirm("acme", []string{"acme/api", "acme/web"}))
	assert.Equal(t, "Delete the workspace labs with its keys and secrets? Its directory goes once nothing else is in it.", view.DeleteWorkspaceConfirm("labs", nil))
	assert.Equal(t, "Delete the default workspace internal with its keys, its home and the projects at the root with their checkouts: apps, mobius? The root's built-in jobs keep running; Add project then needs a workspace.", view.DeleteWorkspaceConfirm("internal", []string{"apps", "mobius"}))
	assert.Equal(t, "Delete the default workspace internal with its keys, its home? The root's built-in jobs keep running; Add project then needs a workspace.", view.DeleteWorkspaceConfirm("internal", nil))
}

// A workspace's page lists its home project's files under ~/, each
// opening in the preview window, and has no Home without one.
func TestFolder_Home(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.Folder = &folder.Folder{ID: "malta", Title: "Malta"}
	p.FolderMenu = &view.FolderMenu{ID: "malta", Workspace: true}
	body := render(t, "folder", p)
	assert.NotContains(t, body, `id="home"`)

	p.HomeFolder = "malta/workspace"
	p.HomeTree = []folder.Item{
		{Name: ".claude", Dir: true, Href: "/f/malta/workspace/.claude", Path: "~/.claude"},
		{Name: "CLAUDE.md", Size: 12, ModTime: time.Now(), Href: "/files/malta/workspace/CLAUDE.md", Path: "~/CLAUDE.md"},
		{Name: ".env", Sealed: true},
	}
	body = render(t, "folder", p)
	assert.Contains(t, body, `<section class="block folder-home" id="home">`)
	assert.Contains(t, body, "malta/workspace/, the $HOME of every run in the workspace")
	assert.Contains(t, body, `<a href="/files/malta/workspace/CLAUDE.md" data-file="~/CLAUDE.md">CLAUDE.md</a>`)
	assert.Contains(t, body, `<a href="/f/malta/workspace/.claude">.claude</a>`)
	assert.Contains(t, body, `title="task-ui lists .env but does not open it"`)
}

// Pin names what the page is a folder of.
func TestFolderMenu_Kind(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "workspace", (&view.FolderMenu{ID: "malta", Workspace: true}).Kind())
	assert.Equal(t, "project", (&view.FolderMenu{ID: "malta/api", Project: true}).Kind())
	assert.Equal(t, "folder", (&view.FolderMenu{ID: "malta/api/web"}).Kind())
}
