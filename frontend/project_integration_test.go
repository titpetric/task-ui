//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestProjects(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/projects")
	assert.Contains(t, body, "<h2>Projects</h2>")
	assert.Contains(t, body, "infra")
	assert.Contains(t, body, "local folder")
	assert.Contains(t, body, `href="/settings/projects/infra"`)
	assert.Contains(t, body, "Root: "+f.root)
	assert.NotContains(t, body, "data-delete") // nothing is a checkout

	form := f.page(t, asAdmin, "/settings/projects/new")
	assert.Contains(t, form, `data-api="/api/projects"`)
	assert.Contains(t, form, "Repository URL")
}

func TestProjectEdit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	empty := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, empty, "<h2>Configure infra</h2>")
	assert.Contains(t, empty, `data-api="/api/projects/infra"`)
	// Who can run is a box a role and a user; none checked is everyone.
	assert.Contains(t, empty, `<input type="checkbox" name="allow" value="runner" data-type="set">`)
	assert.Contains(t, empty, `<input type="checkbox" name="allow" value="ops" data-type="set">`)
	assert.NotContains(t, empty, `name="allow" value="viewer"`)

	project := &model.Project{Folder: "infra", Title: "Infra", Allow: model.EncodeList([]string{"ops", "release"}), Sources: model.EncodeList([]string{"Makefile"})}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	filled := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, filled, `value="Infra"`)
	assert.Contains(t, filled, `<input type="checkbox" name="allow" value="ops" data-type="set" checked>`)
	assert.Contains(t, filled, `<input type="checkbox" name="allow" value="release" data-type="set" checked>`)
	assert.Contains(t, filled, `value="Makefile"`)

	// A folder without sources still opens its settings; a bad name does not.
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/settings/projects/fresh").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/projects/..").Status)
}

func TestProjectEdit_Hours(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	project := &model.Project{Folder: "infra", Hours: "mon-fri 9-17\nfri 9-13"}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	body := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, body, `name="hours"`)
	assert.Contains(t, body, ">mon-fri 9-17\nfri 9-13</textarea>")
}

// The configure page takes who may run and the hours per job, filled
// from the row, as "map.key" fields the form turns into objects.
func TestProjectEdit_PerJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	fresh := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, fresh, `<input type="checkbox" name="job_allow.plan" value="ops" data-type="set">`)
	assert.Contains(t, fresh, `name="job_hours.slow" value=""`)

	project := &model.Project{Folder: "infra", JobAllow: model.EncodeMap(map[string][]string{"plan": {"ops", "release"}}), JobHours: `{"slow":"mon-fri 9-17"}`}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	filled := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, filled, `<input type="checkbox" name="job_allow.plan" value="ops" data-type="set" checked>`)
	assert.Contains(t, filled, `<input type="checkbox" name="job_allow.plan" value="release" data-type="set" checked>`) // no such user: kept
	assert.Contains(t, filled, `name="job_hours.slow" value="mon-fri 9-17"`)
}

// The clone and configure forms have the recursive toggle, on unless
// the project is shallow.
func TestProject_RecursiveToggle(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	add := f.page(t, asAdmin, "/settings/projects/new")
	assert.Contains(t, add, `type="checkbox" id="p-recursive" name="recursive" checked`)

	fresh := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, fresh, `type="checkbox" id="pe-recursive" name="recursive" checked`)

	project := &model.Project{Folder: "infra", Shallow: true}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	shallow := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, shallow, `type="checkbox" id="pe-recursive" name="recursive">`)
	assert.NotContains(t, shallow, `name="recursive" checked`)
}

// The clone form always has a Workspace select, the default one, _,
// chosen, and lists the named ones after it.
func TestProjectNew_Workspace(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	alone := f.page(t, asAdmin, "/settings/projects/new")
	assert.Contains(t, alone, `<select id="p-workspace" name="workspace">`)
	assert.Contains(t, alone, `<option value="_" selected>`)

	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta"}))
	form := f.page(t, asAdmin, "/settings/projects/new")
	assert.Equal(t, 1, strings.Count(form, `<option value="_"`))
	assert.Contains(t, form, `<option value="_" selected>`)
	assert.Contains(t, form, `<option value="malta">malta · Malta</option>`)
}

// A project at the top of a workspace gets a Workspace select on its
// configure page, on the workspace that holds it, naming what a move leaves
// behind; a folder inside a project gets none.
func TestProjectEdit_Move(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan", Expression: "@hourly"}))

	body := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, body, `data-move="infra" data-move-from="_" data-move-left="schedule infra · plan"`)
	assert.Contains(t, body, `<option value="_" selected>`)
	assert.Contains(t, body, `<option value="acme">acme · Acme</option>`)

	assert.NotContains(t, f.page(t, asAdmin, "/settings/projects/infra/edge"), `data-move=`)
	assert.NotContains(t, f.page(t, asAdmin, "/settings/projects/acme"), `data-move=`)
}
