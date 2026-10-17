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
	assert.Contains(t, body, `<option value="private/deploy"`)
	assert.Contains(t, body, `<option value="infra/plan"`)

	ops := f.page(t, asOps, "/schedules?folder=infra")
	assert.Contains(t, ops, `<option value="infra/plan" selected>`)
	assert.NotContains(t, ops, `<option value="private/deploy"`)

	viewer := f.page(t, asViewer, "/schedules")
	assert.NotContains(t, viewer, `data-api="/api/schedules"`)
	assert.Contains(t, viewer, "<code>@hourly</code>")

	limited := f.page(t, asLimited, "/schedules")
	assert.NotContains(t, limited, "<code>0 3 * * *</code>")
}
