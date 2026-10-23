//go:build integration

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
	assert.Contains(t, body, `<code class="nowrap">@hourly</code>`)
	assert.Contains(t, body, `<code class="nowrap">0 3 * * *</code>`)
	// The table fits its box: Last run goes on a narrow one, and a
	// narrower one shows each schedule as a card, State kept.
	assert.Contains(t, body, `<div class="table-wrap fit">
<table class="table fit-cards">`)
	assert.Contains(t, body, `<th class="hide-narrow">Last run</th>`)
	assert.Contains(t, body, `<td data-label="State"><span class="badge ok">enabled</span></td>`)
	assert.Contains(t, body, `class="badge">disabled`)
	assert.Contains(t, body, `<option value="deploy" data-folder="private">`)
	assert.Contains(t, body, `<option value="plan" data-folder="infra">`)

	ops := f.page(t, asOps, "/schedules?folder=infra")
	assert.Contains(t, ops, `<option value="infra" selected>infra</option>`)
	assert.NotContains(t, ops, `<option value="deploy" data-folder="private">`)

	viewer := f.page(t, asViewer, "/schedules")
	assert.NotContains(t, viewer, `data-api="/api/schedules"`)
	assert.Contains(t, viewer, `<code class="nowrap">@hourly</code>`)

	limited := f.page(t, asLimited, "/schedules")
	assert.NotContains(t, limited, `<code class="nowrap">0 3 * * *</code>`)
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

// An admin picks who a schedule runs as, in the form and on each row;
// everyone else schedules as themselves and sees who a schedule runs as.
func TestSchedules_RunAs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan", UserID: f.users[asOps].ID, UserName: asOps, Expression: "@hourly", Enabled: 1}))

	admin := f.page(t, asAdmin, "/schedules")
	assert.Contains(t, admin, `<label for="sch-user">Run as</label>`)
	assert.Contains(t, admin, `<option value="admin" selected>admin (admin)</option>`)
	assert.Contains(t, admin, `<select name="user" data-submit aria-label="Run as">`)
	assert.Contains(t, admin, `<option value="ops" selected>ops</option>`)

	job := f.page(t, asAdmin, "/f/_/infra/plan")
	assert.Contains(t, job, `<label for="sch-user">Run as</label>`)

	for _, path := range []string{"/schedules", "/f/_/infra/plan"} {
		ops := f.page(t, asOps, path)
		assert.NotContains(t, ops, `Run as</label>`, path)
		assert.NotContains(t, ops, `data-submit`, path)
		assert.Contains(t, ops, `<span class="user-name">ops</span>`, path)
	}
}
