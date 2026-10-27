package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// workspacesRow returns the row of the workspace in the rendered list.
func workspacesRow(t *testing.T, body, folder string) string {
	t.Helper()
	start := strings.Index(body, `<a href="/f/`+folder+`">`)
	assert.True(t, start >= 0, "a row for "+folder)
	end := strings.Index(body[start:], "</tr>")
	return body[start : start+end]
}

// The row of a workspace adds a project, lists its keys and its secrets
// without opening Edit workspace first.
func TestWorkspaces_RowMenu(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.WorkspaceList = []model.Workspace{{Folder: "malta", Title: "Malta"}}
	p.Workspaces = map[string]int{"malta": 0}
	row := workspacesRow(t, render(t, "workspaces", p), "malta")
	assert.Contains(t, row, `href="/settings/projects/new?workspace=malta"`)
	assert.Contains(t, row, `href="/settings/workspaces/malta/keys"`)
	assert.Contains(t, row, `href="/settings/secrets?workspace=malta"`)
	add := strings.Index(row, "Add project")
	keys := strings.Index(row, "SSH keys")
	secrets := strings.Index(row, "<span>Secrets</span>")
	edit := strings.Index(row, "Edit workspace")
	assert.True(t, edit < add && add < keys && keys < secrets, "Edit workspace, Add project, SSH keys, Secrets")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	row = workspacesRow(t, render(t, "workspaces", p), "malta")
	assert.NotContains(t, row, "Add project", "adding a project takes project/edit")
	assert.NotContains(t, row, "<span>Secrets</span>", "the secrets take secret/view")
	assert.Contains(t, row, "SSH keys")
}

// A workspace with no project, or no secret, says so and links the form
// that adds one in the cell, for a user who may; with some, the count links.
func TestWorkspaces_EmptyCounts(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.WorkspaceList = []model.Workspace{{Folder: "empty"}, {Folder: "full"}}
	p.Workspaces = map[string]int{"empty": 0, "full": 1}
	p.Nested = map[string][]string{"full": {"full/api", "full/web"}}
	p.SecretCounts = map[string]int{"full": 2}
	body := render(t, "workspaces", p)

	empty := workspacesRow(t, body, "empty")
	assert.Contains(t, empty, `<a href="/settings/projects/new?workspace=empty" aria-label="Add a project to empty">add</a>`)
	assert.Contains(t, empty, `<a href="/settings/secrets?workspace=empty" aria-label="Add a secret to empty">add</a>`)

	full := workspacesRow(t, body, "full")
	assert.Contains(t, full, `<a href="/settings/workspaces/full" title="full/api, full/web">2</a>`)
	assert.NotContains(t, full, "aria-label=\"Add a project")
	assert.NotContains(t, full, "aria-label=\"Add a secret")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	empty = workspacesRow(t, render(t, "workspaces", p), "empty")
	assert.NotContains(t, empty, "aria-label=\"Add a")
	assert.Contains(t, empty, `<span class="faint">none</span>`)
}
