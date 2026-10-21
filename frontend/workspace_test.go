package frontend_test

import (
	"context"
	"net/http"
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

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta", SourceURL: "git@example.com:org/workspace.git", Branch: "main"}))
	body := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, body, `href="/f/malta"`)
	assert.Contains(t, body, "Malta")
	assert.Contains(t, body, "git@example.com:org/workspace.git")
	assert.Contains(t, body, `href="/settings/workspaces/malta"`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta"`)

	settings := f.page(t, asAdmin, "/settings")
	assert.Contains(t, settings, `href="/settings/workspaces"`)
	assert.Contains(t, settings, "1 workspace<")

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
	confirm := "and its checkout, with the projects inside it: malta/deploy, malta/dockerfiles?"
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
