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
