//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestWorkspaces(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	empty := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, empty, "<h2>Workspaces</h2>")
	assert.Contains(t, empty, "No workspaces yet.")
	assert.Contains(t, empty, `href="/settings/workspaces/new"`)
	assert.Contains(t, empty, `<form class="form" data-import="/api/workspaces/import">`)
	assert.Contains(t, empty, `type="file" id="wi-bundle" name="bundle"`)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta", HomeURL: "git@example.com:org/workspace.git"}))
	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `href="/f/malta"`)
	assert.Contains(t, body, "Malta")
	assert.Contains(t, body, `<th>Home repository</th>`)
	assert.Contains(t, body, `>git@example.com:org/workspace.git</span>`, "the home repository lists")
	assert.Contains(t, body, `href="/settings/workspaces/malta"`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta"`)
	// One workspace has nowhere to move.
	assert.NotContains(t, body, `data-workspace-move`)

	settings := f.page(t, asAdmin, "/settings")
	assert.Contains(t, settings, `href="/settings/workspaces"`)
	assert.Contains(t, settings, "1 workspace<")

	// The workspaces list in their order, each movable toward the other.
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "gozo", Title: "Gozo"}))
	body = f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `<span class="faint">empty</span>`, "a workspace without a home repository")
	assert.True(t, strings.Index(body, `href="/f/malta"`) < strings.Index(body, `href="/f/gozo"`), "malta, added first, lists first")
	assert.NotContains(t, body, `data-workspace-move="up" data-folder="malta"`)
	assert.Contains(t, body, `data-workspace-move="down" data-folder="malta"><svg class="ic"><use href="#i-down"/></svg><span>Move down</span>`)
	assert.Contains(t, body, `data-workspace-move="up" data-folder="gozo"><svg class="ic"><use href="#i-up"/></svg><span>Move up</span>`)
	assert.NotContains(t, body, `data-workspace-move="down" data-folder="gozo"`)

	form := f.page(t, asAdmin, "/settings/workspaces/new")
	assert.Contains(t, form, `data-api="/api/workspaces"`)
	assert.Contains(t, form, `name="folder"`)
	assert.Contains(t, form, `name="url"`)
	assert.Contains(t, form, `name="branch"`)
	assert.NotContains(t, form, `name="project"`)
}

func TestWorkspaceEdit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta", Description: "The org"}))
	body := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, body, "<h2>Edit malta</h2>")
	assert.Contains(t, body, `data-api="/api/workspaces/malta" data-method="PUT"`)
	assert.Contains(t, body, `value="Malta"`)
	assert.Contains(t, body, `value="The org"`)
	assert.Contains(t, body, "The home is an empty directory.")
	assert.Contains(t, body, `type="checkbox" id="we-recursive" name="recursive" checked`)
	assert.Contains(t, f.page(t, asAdmin, "/settings/workspaces/new"), `type="checkbox" id="w-recursive" name="recursive" checked`)

	// The Sidebar field: empty for every project, the number when set.
	assert.Contains(t, body, `<input type="number" id="we-sidebar" name="sidebar_limit" data-type="int" min="0" step="1" placeholder="all" value="" autocomplete="off">`)
	assert.Contains(t, body, "How many projects the sidebar lists for this workspace, the most recently committed first. Empty or 0 lists them all, by name.")

	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "gozo", Shallow: true, SidebarLimit: 4}))
	shallow := f.page(t, asAdmin, "/settings/workspaces/gozo")
	assert.Contains(t, shallow, `type="checkbox" id="we-recursive" name="recursive">`)
	assert.Contains(t, shallow, "off lists only the folder and its top-level folders")
	assert.Contains(t, shallow, `name="sidebar_limit" data-type="int" min="0" step="1" placeholder="all" value="4"`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/nope").Status)
}

// The delete confirm names the projects inside the workspace, the rows
// whose folder starts with it, and no other.
func TestWorkspaceEdit_DeleteNamesProjects(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta"}))
	for _, folder := range []string{"malta/deploy", "malta/dockerfiles", "maltese", "infra"} {
		assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: folder}))
	}
	confirm := "with its keys and secrets, and the projects inside it with their checkouts: malta/deploy, malta/dockerfiles?"
	for _, path := range []string{"/settings/workspaces/malta", "/settings/workspaces"} {
		body := f.page(t, asAdmin, path)
		assert.Contains(t, body, confirm)
		assert.NotContains(t, body, "maltese")
	}
	edit := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, edit, `data-next="/settings/workspaces?notice=Workspace+deleted"`)
}

func TestWorkspaces_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: "malta"}))
	for _, path := range []string{"/settings/workspaces", "/settings/workspaces/new", "/settings/workspaces/malta"} {
		for _, user := range []string{asOps, asViewer} {
			res := f.get(t, user, path)
			assert.Equal(t, http.StatusForbidden, res.Status, path)
			assert.Contains(t, res.Body, "You need permission to open this page.")
		}
	}
}

// The default workspace lists first, links to /f/_, does not move and
// deletes like any other; the first named workspace has nowhere up to go.
func TestWorkspaces_Default(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta"}))
	assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: "site"}))
	// A checkout under src counts with or without a project row; the
	// workspace's checkouts count under it, a directory that is no
	// checkout does not.
	for _, dir := range []string{"src/site", "src/extra", "malta/src/deploy/.git", "malta/src/notes"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(f.root, dir), 0o755))
	}

	f.withKey(t)
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:malta", Name: "TOKEN", Content: "shared value"}))

	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.True(t, strings.Index(body, `href="/f/internal"`) < strings.Index(body, `href="/f/malta"`), "internal lists first")
	assert.Contains(t, body, `<th class="hide-sm">Secrets</th>`)
	assert.Contains(t, body, `<a href="/settings/workspaces/malta/secrets" title="delete them before deleting the workspace">1</a>`)
	assert.Contains(t, body, `href="/settings/workspaces/malta/secrets"><svg class="ic"><use href="#i-secret"/></svg><span>Delete its 1 secret first</span>`)
	assert.NotContains(t, body, `data-delete="/api/workspaces/malta"`, "a workspace with secrets has no Delete")
	assert.Contains(t, body, `<a href="/settings/workspaces/internal" title="extra, infra, private, site">4</a>`)
	assert.Contains(t, body, `<a href="/settings/workspaces/malta" title="malta/deploy">1</a>`)
	assert.Contains(t, body, `data-delete="/api/workspaces/internal"`)
	assert.Contains(t, body, `Delete the default workspace internal with its keys, its home and the projects at the root with their checkouts: extra, infra, private, site? The root&#39;s built-in jobs keep running; Add project then needs a workspace.`)
	assert.NotContains(t, body, `Empty</span>`)
	assert.NotContains(t, body, `data-folder="internal"`)
	assert.NotContains(t, body, `data-workspace-move="up" data-folder="malta"`)
	edit := f.page(t, asAdmin, "/settings/workspaces/internal")
	assert.Contains(t, edit, `data-api="/api/workspaces/internal"`)
	assert.Contains(t, edit, "the root, "+f.root)
	assert.Contains(t, edit, `data-delete="/api/workspaces/internal"`)
	assert.Contains(t, edit, `data-next="/settings/workspaces?notice=Workspace+deleted"`)
	assert.NotContains(t, edit, "emptied")

	// A workspace with nothing in it.
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "gozo"}))
	body = f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `<span class="faint">none</span>`)
	assert.Contains(t, body, `data-delete="/api/workspaces/gozo"`)

	malta := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, malta, `<span>Delete its 1 secret first</span>`)
	assert.NotContains(t, malta, `data-delete="/api/workspaces/malta"`)
}

// An instance without the default workspace, which deleted it, lists
// the named workspaces alone: no _ on the Workspaces page, the clone
// form, the Projects page or the Secrets page, and nothing says Empty;
// /f/_ still renders, the root's page, with no workspace menu.
func TestWorkspaces_NoDefault(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta"}))
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, "malta", "src", "deploy", ".git"), 0o755))
	f.withKey(t)

	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `href="/f/malta"`)
	main := body[strings.Index(body, "<main"):]
	assert.NotContains(t, main, `href="/f/internal"`, "the header's prompt still names the root; the list does not")
	assert.NotContains(t, main, "Empty")
	assert.Contains(t, body, `data-delete="/api/workspaces/malta"`)
	assert.Contains(t, body, `data-confirm-ok="Delete"`)

	projects := f.page(t, asAdmin, "/settings/projects")
	assert.Contains(t, projects, "inside malta/src/")
	assert.NotContains(t, projects, "in "+f.root+"/src")

	secrets := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, secrets, `scope=workspace:malta`)
	assert.NotContains(t, secrets, `scope=workspace:internal`)

	root := f.get(t, asAdmin, "/f/internal")
	assert.Equal(t, http.StatusOK, root.Status)
	assert.NotContains(t, root.Body, `data-delete="/api/workspaces/internal"`)
	assert.NotContains(t, root.Body, `/api/workspaces/internal/prompt`)
}
