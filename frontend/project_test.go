package frontend_test

import (
	"context"
	"net/http"
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
	assert.Contains(t, empty, `name="allow" data-type="list" value=""`)

	project := &model.Project{Folder: "infra", Title: "Infra", Allow: model.EncodeList([]string{"ops", "release"}), Sources: model.EncodeList([]string{"Makefile"})}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	filled := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, filled, `value="Infra"`)
	assert.Contains(t, filled, `value="ops, release"`)
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
	assert.Contains(t, fresh, `name="job_allow.plan" data-type="list" value=""`)
	assert.Contains(t, fresh, `name="job_hours.slow" value=""`)

	project := &model.Project{Folder: "infra", JobAllow: model.EncodeMap(map[string][]string{"plan": {"ops", "release"}}), JobHours: `{"slow":"mon-fri 9-17"}`}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	filled := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, filled, `name="job_allow.plan" data-type="list" value="ops, release"`)
	assert.Contains(t, filled, `name="job_hours.slow" value="mon-fri 9-17"`)
}

// The clone and configure forms have the recursive toggle, on unless
// the project is root-only.
func TestProject_RecursiveToggle(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	add := f.page(t, asAdmin, "/settings/projects/new")
	assert.Contains(t, add, `type="checkbox" id="p-recursive" name="recursive" checked`)

	fresh := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, fresh, `type="checkbox" id="pe-recursive" name="recursive" checked`)

	project := &model.Project{Folder: "infra", RootOnly: true}
	assert.NoError(t, f.opts.Projects.Create(context.Background(), project))
	rootOnly := f.page(t, asAdmin, "/settings/projects/infra")
	assert.Contains(t, rootOnly, `type="checkbox" id="pe-recursive" name="recursive">`)
	assert.NotContains(t, rootOnly, `name="recursive" checked`)
}
