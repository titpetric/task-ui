//go:build integration

package frontend_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
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

// A workflow's schedule lists as a workflow with the name linked, for
// users who see every step of it; the page's own form stays jobs only.
func TestSchedules_Workflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	wf := f.workflow(t, asOps, "release", "infra", "plan", "infra/edge", "deploy")
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{WorkflowID: wf.ID, UserID: f.users[asOps].ID, UserName: asOps, Expression: "0 7 * * 1-5", Enabled: 1}))

	body := f.page(t, asOps, "/schedules")
	assert.Contains(t, body, "Schedule a workflow from its edit page.")
	assert.Contains(t, body, `<td data-label="Folder"><svg class="ic lead"><use href="#i-flow"/></svg>workflow</td>`)
	assert.Contains(t, body, `<td class="key" data-label="Job"><a href="/workflows/`+wf.ID+`">release</a></td>`)
	assert.Contains(t, body, `data-confirm="Delete the schedule 0 7 * * 1-5 for release?"`)
	assert.NotContains(t, body, `name="workflow"`)

	limited := f.page(t, asLimited, "/schedules")
	assert.NotContains(t, limited, `0 7 * * 1-5`, "limited does not see infra/edge")
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
	// A key linked to the runners' role shows on ops's options, in the
	// form and on the row; the admin, with no key, says so.
	deploy, err := f.opts.Keys.Generate(ctx, keystore.Spec{Name: "deploy"})
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Keys.Link(ctx, deploy.Name, []string{model.RoleRunner}, nil))

	admin := f.page(t, asAdmin, "/schedules")
	assert.Contains(t, admin, `<label for="sch-user">Run as</label>`)
	assert.Contains(t, admin, `<option value="admin" selected data-keys="">admin (admin) · no keys</option>`)
	assert.Contains(t, admin, `<option value="ops" data-keys="deploy_ed25519">ops (runner) · keys: deploy_ed25519</option>`)
	assert.Contains(t, admin, `<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>`)
	assert.Contains(t, admin, `<select name="user" data-submit aria-label="Run as">`)
	assert.Contains(t, admin, `<option value="ops" selected data-keys="deploy_ed25519">ops · keys: deploy_ed25519</option>`)

	job := f.page(t, asAdmin, "/f/_/infra/plan")
	assert.Contains(t, job, `<label for="sch-user">Run as</label>`)
	assert.Contains(t, job, `<option value="ops" data-keys="deploy_ed25519">ops (runner) · keys: deploy_ed25519</option>`)

	for _, path := range []string{"/schedules", "/f/_/infra/plan"} {
		ops := f.page(t, asOps, path)
		assert.NotContains(t, ops, `Run as</label>`, path)
		assert.NotContains(t, ops, `data-submit`, path)
		assert.Contains(t, ops, `<span class="user-name">ops</span>`, path)
	}
}
