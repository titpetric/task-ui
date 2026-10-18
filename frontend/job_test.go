package frontend_test

import (
	"context"
	"net/http"
	"testing"

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
