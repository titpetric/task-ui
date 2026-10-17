package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRuns(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	plan := f.run(t, "infra", "plan", "ops", model.StateOK)
	deploy := f.run(t, "private", "deploy", "admin", model.StateFailed)

	body := f.page(t, asAdmin, "/runs")
	assert.Contains(t, body, "<h2>Runs</h2>")
	assert.Contains(t, body, `href="/runs/`+plan.ID+`"`)
	assert.Contains(t, body, `href="/runs/`+deploy.ID+`"`)

	// The sidebar lists every recent run; the table only the filtered ones.
	filtered := f.page(t, asAdmin, "/runs?folder=infra&state=ok")
	assert.Contains(t, filtered, `class="btn btn-sm" href="/runs/`+plan.ID+`"`)
	assert.NotContains(t, filtered, `class="btn btn-sm" href="/runs/`+deploy.ID+`"`)
	assert.Contains(t, filtered, `<option value="infra" selected>`)
	assert.Contains(t, filtered, `<option value="ok" selected>`)

	limited := f.page(t, asLimited, "/runs")
	assert.NotContains(t, limited, deploy.ID)
}

func TestRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "plan", "ops", model.StateFailed)
	run.ExitCode, run.CastPath = 2, "infra/plan/x.cast"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))
	assert.NoError(t, f.opts.Artifacts.Create(context.Background(), &model.Artifact{RunID: run.ID, Folder: "infra", Job: "plan", Name: "report.txt", Path: "x/report.txt", ByteSize: 2048, ContentType: "text/plain"}))

	body := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, body, "<h2>infra / plan")
	assert.Contains(t, body, "failed")
	assert.Contains(t, body, `data-attach="`+run.ID+`">View</button>`)
	assert.Contains(t, body, `href="/api/runs/`+run.ID+`/cast"`)
	assert.Contains(t, body, "report.txt")
	assert.Contains(t, body, "2.0 KB")
	assert.NotContains(t, body, "data-frame")
	assert.NotContains(t, body, "data-delete")

	admin := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, admin, `data-delete="/api/artifacts/`)
}

func TestRun_Live(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/runs/"+session.ID())
	assert.Contains(t, body, `data-attach="`+session.ID()+`">Attach</button>`)
	assert.Contains(t, body, `data-stop="`+session.ID()+`"`)
	assert.Contains(t, body, `data-kill="`+session.ID()+`"`)
	assert.Contains(t, body, `data-frame="`+session.ID()+`"`)
	assert.Contains(t, body, "nobody attached")

	viewer := f.page(t, asViewer, "/runs/"+session.ID())
	assert.NotContains(t, viewer, "data-stop")
	assert.Contains(t, viewer, "data-frame")
}

func TestRun_Missing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	hidden := f.run(t, "private", "deploy", "admin", model.StateOK)

	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/runs/"+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/runs/"+hidden.ID).Status)
}
