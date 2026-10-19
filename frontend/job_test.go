package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "plan", "ops", model.StateOK)
	f.run(t, "infra", "slow", "ops", model.StateKilled)
	assert.NoError(t, f.opts.Schedules.Create(context.Background(), &model.Schedule{Folder: "infra", Job: "plan", UserName: "ops", Expression: "@hourly", Enabled: 1}))

	body := f.page(t, asOps, "/f/infra/plan")
	assert.Contains(t, body, "<h2>plan")
	// A script is its own source: there is no pipeline file to show.
	assert.NotContains(t, body, `class="btn btn-sm show-source"`)
	assert.Contains(t, body, "<code>sh -c echo planned</code>")
	assert.Contains(t, body, "1 runs")
	assert.Contains(t, body, `data-attach="`+run.ID+`">View</a>`)
	assert.Contains(t, body, "<code>@hourly</code>")
	assert.Contains(t, body, `data-api="/api/schedules"`)
	assert.Contains(t, body, "admin, runner")

	viewer := f.page(t, asViewer, "/f/infra/plan")
	assert.Contains(t, viewer, "is not allowed to run it")
	assert.NotContains(t, viewer, `data-api="/api/schedules"`)
}

func TestJob_Running(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/f/infra/slow")
	assert.Contains(t, body, "<b>Running</b>")
	assert.Contains(t, body, `data-attach="`+session.ID()+`">Attach</a>`)
	assert.Contains(t, body, `data-kill="`+session.ID()+`"`)
}

func TestJob_Missing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.get(t, asOps, "/f/infra/nothing")
	assert.Equal(t, http.StatusNotFound, res.Status)
	assert.Contains(t, res.Body, "No job named nothing in infra")
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/nowhere/plan").Status)
}

// A pipeline's job shows its file and the files it includes, each with
// a Show button; the folder page does too.
func TestJob_ShowSource(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.infra.Sources = []*detector.Source{{Kind: "atkins", File: "atkins.yml"}}
	f.infra.Includes = []string{"vars.yml"}
	f.infra.Job("plan").Kind, f.infra.Job("plan").Source = "atkins", "atkins.yml"

	job := f.page(t, asOps, "/f/infra/plan")
	assert.Contains(t, job, `atkins.yml<button type="button" class="btn btn-sm show-source" data-source="infra" data-file="atkins.yml" aria-haspopup="dialog">`)
	assert.Contains(t, job, `data-source="infra" data-file="vars.yml"`)
	assert.Contains(t, job, `<dialog class="run source" id="source-dialog"`)

	folder := f.page(t, asOps, "/f/infra")
	assert.Contains(t, folder, `data-source="infra" data-file="atkins.yml"`)
	assert.Contains(t, folder, `data-source="infra" data-file="vars.yml"`)
}
