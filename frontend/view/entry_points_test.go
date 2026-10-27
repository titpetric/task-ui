package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// sideMenu returns the top nav of the sidebar.
func sideMenu(t *testing.T, body string) string {
	t.Helper()
	at := strings.Index(body, `<nav class="menu">`)
	assert.True(t, at >= 0, "the sidebar menu")
	rest := body[at:]
	return rest[:strings.Index(rest, "</nav>")]
}

// Settings is in the sidebar, after Notes, for a user who may open any
// settings page; it sits under Settings, so it is active on its pages.
func TestSidebar_Settings(t *testing.T) {
	t.Parallel()

	p := admin(t)
	menu := sideMenu(t, render(t, "home", p))
	notes, settings := strings.Index(menu, `href="/notes"`), strings.Index(menu, `<a href="/settings" class="">`)
	assert.True(t, notes > 0 && settings > notes, "Settings after Notes")

	p.Path = "/settings/keys"
	assert.Contains(t, sideMenu(t, render(t, "home", p)), `<a href="/settings" class="active">`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	assert.NotContains(t, sideMenu(t, render(t, "home", p)), `href="/settings"`, "nothing to open, no entry")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow})
	assert.Contains(t, sideMenu(t, render(t, "home", p)), `href="/settings"`)
}

// With no folder the sidebar says so and links the form that adds a
// repository, for a user who may; others read where it is added.
func TestSidebar_EmptyLinksAddProject(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.SideFolders, p.Folders = nil, nil
	body := render(t, "home", p)
	assert.Contains(t, body, `No folders yet. <a href="/settings/projects/new">Add a git repository</a>, or put a folder with jobs under the root.`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	body = render(t, "home", p)
	assert.Contains(t, body, `No folders yet. Add a git repository in Settings, or put a folder with jobs under the root.`)
	assert.NotContains(t, body, `href="/settings/projects/new"`)
}

// The dashboard's head adds anything the user may, behind one menu; a
// user who may add nothing sees no menu.
func TestHome_AddMenu(t *testing.T) {
	t.Parallel()

	p := admin(t)
	head := pageHead(t, render(t, "home", p))
	for _, label := range []string{"Add workspace", "Add project", "Generate key", "Upload key", "Add secret", "Add workflow"} {
		assert.Contains(t, head, "<span>"+label+"</span>")
	}
	assert.NotContains(t, head, "Add user", "the server writes no password file here")
	p.CanAddUsers = true
	assert.Contains(t, pageHead(t, render(t, "home", p)), `href="/settings/users/new"`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	assert.NotContains(t, pageHead(t, render(t, "home", p)), "more-menu")
}

// The keyboard's first stop is a link past the sidebar to the content,
// which takes the focus; the header's prompt and its logo button sit on
// the Tab path.
func TestLayout_Keyboard(t *testing.T) {
	t.Parallel()

	body := render(t, "home", admin(t))
	skip := strings.Index(body, `<a class="skip-link" href="#main">Skip to content</a>`)
	header := strings.Index(body, `<header class="top"`)
	main := strings.Index(body, `<main id="main" tabindex="-1">`)
	assert.True(t, skip > 0 && skip < header && header < main, "skip link, header, then main")
	head := body[header:strings.Index(body, `</header>`)]
	assert.NotContains(t, head, `tabindex="-1"`)
	assert.Contains(t, head, `id="header-toggle" aria-label="Prompt"`)
	assert.Contains(t, head, `id="header-prompt" aria-label="Ask Claude"`)

	login := &view.Page{Title: "Sign in"}
	assert.NotContains(t, render(t, "login", login), "skip-link", "no sidebar on the sign in screen")
}

// Add workspace lands on the new workspace, where its project and its
// keys are one click away, not on the list.
func TestWorkspaceNew_LandsOnTheWorkspace(t *testing.T) {
	t.Parallel()

	body := render(t, "workspace_new", admin(t))
	assert.Contains(t, body, `data-api="/api/workspaces" data-method="POST" data-next="/settings/workspaces/{folder}?notice=Workspace+added"`)
}
