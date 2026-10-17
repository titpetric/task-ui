package frontend_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestFolder(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.run(t, "infra", "plan", "ops", model.StateFailed)

	body := f.page(t, asOps, "/f/infra")
	assert.Contains(t, body, "The servers")
	assert.Contains(t, body, `data-run-all data-folder="infra"`)
	assert.Contains(t, body, `data-run data-folder="infra" data-job="plan"`)
	assert.Contains(t, body, "Run force")
	assert.Contains(t, body, `<i class="dot failed"></i>failed`)
	assert.Contains(t, body, ".env")
	assert.Contains(t, body, `class="badge bad">missing`)
	assert.NotContains(t, body, "/f/infra/env/.env") // admins edit env files

	viewer := f.page(t, asViewer, "/f/infra")
	assert.NotContains(t, viewer, "data-run-all")
	assert.NotContains(t, viewer, "data-force")
	assert.Contains(t, viewer, "History</a>")

	admin := f.page(t, asAdmin, "/f/infra")
	assert.Contains(t, admin, `href="/f/infra/env/.env"`)
}

func TestFolder_Running(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/f/infra")
	assert.Contains(t, body, "running, ops")
	assert.Contains(t, body, `data-attach="`+session.ID()+`"`)
	assert.Contains(t, body, `data-kill="`+session.ID()+`"`)

	viewer := f.page(t, asViewer, "/f/infra")
	assert.Contains(t, viewer, `data-attach="`+session.ID()+`"`)
	assert.NotContains(t, viewer, "data-kill")
}

func TestFolder_Missing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/nowhere").Status)
	hidden := f.get(t, asLimited, "/f/private")
	assert.Equal(t, http.StatusNotFound, hidden.Status)
	assert.Contains(t, hidden.Body, "No folder named private")
}

// The sidebar opens the folder of the page into its jobs, a menu button
// drives the drawer, and the jobs table folds into cards.
func TestFolder_SidebarAndCards(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/f/infra")
	assert.Contains(t, body, `<details class="folder active" data-folder="infra" open>`)
	// Few jobs: every folder opens; the sidebar runs a job in place and
	// attaches to the one running.
	assert.Contains(t, body, `<details class="folder" data-folder="private" open>`)
	assert.Contains(t, body, `<div class="job" data-job-key="infra/plan">`)
	assert.Contains(t, body, `class="play" data-run data-folder="infra" data-job="plan"`)
	assert.Contains(t, body, `<div class="job" data-job-key="infra/slow">
							<i class="dot running"></i>`)
	assert.Contains(t, body, `class="play" data-attach="`+session.ID()+`"`)
	assert.NotContains(t, body, `class="play" data-run data-folder="private"`) // ops may not run private
	assert.Contains(t, body, `id="menu-toggle"`)
	assert.Contains(t, body, `id="menu-backdrop"`)
	assert.Contains(t, body, `<div class="jobcards">`)
	assert.Contains(t, body, `<article class="jobcard">`)
	assert.Contains(t, body, `<article class="jobcard running">`)
	assert.Contains(t, body, "<th>Hours</th><td><span class=\"faint\">always</span></td>")

	job := f.page(t, asOps, "/f/infra/plan")
	assert.Contains(t, job, `<div class="job active" data-job-key="infra/plan">`)

	viewer := f.page(t, asViewer, "/f/infra")
	assert.NotContains(t, viewer, `class="play" data-run`)
	assert.Contains(t, viewer, `class="play" data-attach="`+session.ID()+`"`)
}

func TestFolder_Hours(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.infra.Hours = "mon 0:00-0:01"
	f.infra.Job("plan").Hours = "daily 0-24"

	body := f.page(t, asOps, "/f/infra")
	assert.Contains(t, body, "mon 00:00-00:01 <span class=\"badge bad\">out of hours now</span>")

	job := f.page(t, asOps, "/f/infra/plan")
	assert.Contains(t, job, "daily 0-24 <span class=\"badge ok\">open now</span>")
	assert.Contains(t, job, "the folder runs mon 00:00-00:01")

	slow := f.page(t, asOps, "/f/infra/slow")
	assert.Contains(t, slow, "out of hours now")
}
