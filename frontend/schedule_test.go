package frontend_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSchedules(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan", UserName: "ops", Expression: "@hourly", Enabled: 1}))
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{Folder: "private", Job: "deploy", UserName: "admin", Expression: "0 3 * * *", Enabled: 0}))

	body := f.page(t, asAdmin, "/schedules")
	assert.Contains(t, body, "<h2>Scheduled</h2>")
	assert.Contains(t, body, "<code>@hourly</code>")
	assert.Contains(t, body, "<code>0 3 * * *</code>")
	assert.Contains(t, body, `class="badge">disabled`)
	assert.Contains(t, body, `<option value="deploy" data-folder="private">`)
	assert.Contains(t, body, `<option value="plan" data-folder="infra">`)

	ops := f.page(t, asOps, "/schedules?folder=infra")
	assert.Contains(t, ops, `<option value="infra" selected>infra</option>`)
	assert.NotContains(t, ops, `<option value="deploy" data-folder="private">`)

	viewer := f.page(t, asViewer, "/schedules")
	assert.NotContains(t, viewer, `data-api="/api/schedules"`)
	assert.Contains(t, viewer, "<code>@hourly</code>")

	limited := f.page(t, asLimited, "/schedules")
	assert.NotContains(t, limited, "<code>0 3 * * *</code>")
}

// The folder select narrows the job select; viewers get neither.
func TestSchedules_Selects(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asOps, "/schedules?folder=infra")
	assert.Contains(t, body, `<select id="sch-folder" name="folder" data-narrows="sch-job" required>`)
	assert.Contains(t, body, `<option value="infra" selected>infra</option>`)
	assert.NotContains(t, body, `<option value="private">`)
	assert.Contains(t, body, `<option value="plan" data-folder="infra">plan</option>`)
	assert.NotContains(t, body, `<option value="deploy" data-folder="private">`)

	admin := f.page(t, asAdmin, "/schedules")
	assert.Contains(t, admin, `<option value="deploy" data-folder="private">deploy</option>`)
}
