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
	// The table follows the runs topic and is drawn on, never reloaded.
	assert.Contains(t, body, `<section class="block" data-sse="runs">`)
	assert.Contains(t, body, `<tr data-run-id="`+plan.ID+`">`)
	// Only the View button opens the window: the row is not a Run button,
	// the folder and the job are plain links to their pages.
	assert.NotContains(t, body, `<tr data-run=`)
	assert.Contains(t, body, `<td data-label="Folder"><a href="/f/infra">infra</a></td>`)
	assert.Contains(t, body, `<td class="key"><a href="/f/infra/plan">plan</a></td>`)
	assert.NotContains(t, body, `data-live`)

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
	assert.Contains(t, body, `<h2>infra / plan <span class="badge bad">failed 2</span></h2>`)
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

func TestRun_Reason(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "plan", "ops", model.StateOK)
	run.Forced, run.Reason = 1, "customer waiting on the fix"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))

	page := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, page, "<th>Reason</th><td>customer waiting on the fix")
	assert.Contains(t, page, `title="customer waiting on the fix">out of hours</span>`)

	list := f.page(t, asViewer, "/runs")
	assert.Contains(t, list, `title="customer waiting on the fix">out of hours</span>`)
	assert.NotContains(t, list, ">forced</span>")
	assert.Contains(t, list, `<table class="table cards">`)
	assert.Contains(t, list, `data-label="Result"`)
}

// A run someone watches says so in its result; the sidebar says what a
// click does by permission.
func TestRuns_Watching(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")
	sub := session.Attach(true)
	defer session.Detach(sub)

	ops := f.page(t, asOps, "/runs")
	assert.Contains(t, ops, `running</span> <span class="faint">1 watching</span>`)
	assert.NotContains(t, ops, "<th>Watching</th>")
	assert.Contains(t, ops, `title="Attach to infra / slow, ops, running"`)
	assert.Contains(t, ops, `<svg class="ic act"><use href="#i-play"/></svg>`)

	viewer := f.page(t, asViewer, "/runs")
	assert.Contains(t, viewer, `title="Watch infra / slow, ops, running"`)
	assert.Contains(t, viewer, `<svg class="ic act"><use href="#i-eye"/></svg>`)
}
