package frontend_test

import (
	"context"
	"net/http"
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
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta", SourceURL: "git@example.com:org/workspace.git", Branch: "main"}))
	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `href="/f/malta"`)
	assert.Contains(t, body, "Malta")
	assert.Contains(t, body, "git@example.com:org/workspace.git")
	assert.Contains(t, body, `href="/settings/workspaces/malta"`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta"`)
	// One workspace has nowhere to move.
	assert.Contains(t, body, `data-workspace-move="up" data-folder="malta" title="Move up" aria-label="Move malta up" disabled>`)
	assert.Contains(t, body, `data-workspace-move="down" data-folder="malta" title="Move down" aria-label="Move malta down" disabled>`)

	settings := f.page(t, asAdmin, "/settings")
	assert.Contains(t, settings, `href="/settings/workspaces"`)
	assert.Contains(t, settings, "1 workspace<")

	// The workspaces list in their order, each movable toward the other.
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "gozo", Title: "Gozo"}))
	body = f.page(t, asAdmin, "/settings/workspaces")
	assert.True(t, strings.Index(body, `href="/f/malta"`) < strings.Index(body, `href="/f/gozo"`), "malta, added first, lists first")
	assert.Contains(t, body, `data-workspace-move="up" data-folder="malta" title="Move up" aria-label="Move malta up" disabled>`)
	assert.Contains(t, body, `data-workspace-move="down" data-folder="malta" title="Move down" aria-label="Move malta down">`)
	assert.Contains(t, body, `data-workspace-move="up" data-folder="gozo" title="Move up" aria-label="Move gozo up">`)
	assert.Contains(t, body, `data-workspace-move="down" data-folder="gozo" title="Move down" aria-label="Move gozo down" disabled>`)

	form := f.page(t, asAdmin, "/settings/workspaces/new")
	assert.Contains(t, form, `data-api="/api/workspaces"`)
	assert.Contains(t, form, `name="folder"`)
	assert.Contains(t, form, `name="url"`)
}

func TestWorkspaceEdit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta", Description: "The org", SourceURL: "git@example.com:org/workspace.git"}))
	body := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, body, "<h2>Configure malta</h2>")
	assert.Contains(t, body, `data-api="/api/workspaces/malta" data-method="PUT"`)
	assert.Contains(t, body, `value="Malta"`)
	assert.Contains(t, body, `value="The org"`)
	assert.Contains(t, body, "none yet")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/nope").Status)
}

// The delete confirm names the projects inside the workspace, the rows
// whose folder starts with it, and no other.
func TestWorkspaceEdit_DeleteNamesProjects(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", SourceURL: "git@example.com:org/workspace.git"}))
	for _, folder := range []string{"malta/deploy", "malta/dockerfiles", "maltese", "infra"} {
		assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: folder}))
	}
	confirm := "its checkout, its keys and its secrets, with the projects inside it: malta/deploy, malta/dockerfiles?"
	for _, path := range []string{"/settings/workspaces/malta", "/settings/workspaces"} {
		body := f.page(t, asAdmin, path)
		assert.Contains(t, body, confirm)
		assert.NotContains(t, body, "maltese")
	}
	edit := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, edit, `data-next="/settings/workspaces?notice=Workspace&#43;removed"`)
}

func TestWorkspaces_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: "malta"}))
	for _, path := range []string{"/settings/workspaces", "/settings/workspaces/new", "/settings/workspaces/malta"} {
		for _, user := range []string{asOps, asViewer} {
			res := f.get(t, user, path)
			assert.Equal(t, http.StatusForbidden, res.Status, path)
			assert.Contains(t, res.Body, "Settings are for admins.")
		}
	}
}

// The default workspace lists first, links to /f/_, and neither moves
// nor goes; the first named workspace has nowhere up to go.
func TestWorkspaces_Default(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta"}))
	assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: "site"}))

	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.True(t, strings.Index(body, `href="/f/_"`) < strings.Index(body, `href="/f/malta"`), "_ lists first")
	assert.Contains(t, body, `<a href="/f/_/site">site</a>`)
	assert.NotContains(t, body, `data-delete="/api/workspaces/_"`)
	assert.NotContains(t, body, `data-folder="_" title="Move`)
	assert.Contains(t, body, `data-workspace-move="up" data-folder="malta" title="Move up" aria-label="Move malta up" disabled>`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta"`)

	edit := f.page(t, asAdmin, "/settings/workspaces/_")
	assert.Contains(t, edit, `data-api="/api/workspaces/_"`)
	assert.Contains(t, edit, "the root, "+f.root)
	assert.NotContains(t, edit, "data-delete=")
}
