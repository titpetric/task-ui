package api_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

func ptr[T any](v T) *T { return &v }

func TestSchedule_Create(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asOps, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "0 * * * *"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.True(t, model.ValidID(res.str("id")))
	assert.Equal(t, "ops", res.str("user_name"))
	assert.Equal(t, float64(1), res.get("enabled"))
	assert.NotEqual(t, "", res.str("next_run_at"))

	off := f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@daily", Enabled: ptr(false)})
	assert.Equal(t, http.StatusCreated, off.Status)
	assert.Equal(t, float64(0), off.get("enabled"))
}

func TestSchedule_CreateRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	cases := []struct {
		name   string
		user   string
		req    api.ScheduleRequest
		status int
	}{
		{"viewer", asViewer, api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily"}, http.StatusForbidden},
		{"not allowed", asOps, api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@daily"}, http.StatusForbidden},
		{"no folder", asOps, api.ScheduleRequest{Folder: "nowhere", Job: "plan", Expression: "@daily"}, http.StatusNotFound},
		{"no job", asOps, api.ScheduleRequest{Folder: "infra", Job: "nothing", Expression: "@daily"}, http.StatusNotFound},
		{"a shell", asAdmin, api.ScheduleRequest{Folder: "infra", Job: "@shell", Expression: "@daily"}, http.StatusBadRequest},
		{"bad expression", asOps, api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "every day"}, http.StatusBadRequest},
	}
	for _, c := range cases {
		res := f.call(t, c.user, http.MethodPost, "/schedules", c.req)
		assert.Equal(t, c.status, res.Status, c.name)
	}
}

func TestSchedule_ListAndGet(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	plan := f.call(t, asOps, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@hourly"})
	f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@daily"})

	all := f.call(t, asViewer, http.MethodGet, "/schedules", nil)
	assert.Equal(t, http.StatusOK, all.Status)
	assert.Len(t, all.list("schedules"), 2)

	limited := f.call(t, asLimited, http.MethodGet, "/schedules", nil)
	assert.Len(t, limited.list("schedules"), 1)

	one := f.call(t, asViewer, http.MethodGet, "/schedules/"+plan.str("id"), nil)
	assert.Equal(t, http.StatusOK, one.Status)
	assert.Equal(t, "@hourly", one.str("expression"))

	missing := f.call(t, asViewer, http.MethodGet, "/schedules/"+model.NewID(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

func TestSchedule_Update(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asOps, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@hourly"})
	path := "/schedules/" + created.str("id")

	denied := f.call(t, asViewer, http.MethodPut, path, api.ScheduleRequest{Expression: "@daily"})
	assert.Equal(t, http.StatusForbidden, denied.Status)

	bad := f.call(t, asOps, http.MethodPut, path, api.ScheduleRequest{Expression: "sometimes"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	res := f.call(t, asOps, http.MethodPut, path, api.ScheduleRequest{Expression: "@daily", Enabled: ptr(false)})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, "@daily", res.str("expression"))
	assert.Equal(t, float64(0), res.get("enabled"))

	// Another runner of the job may switch it back on.
	on := f.call(t, asLimited, http.MethodPut, path, api.ScheduleRequest{Enabled: ptr(true)})
	assert.Equal(t, http.StatusOK, on.Status)
	assert.Equal(t, float64(1), on.get("enabled"))

	missing := f.call(t, asOps, http.MethodPut, "/schedules/"+model.NewID(), api.ScheduleRequest{Enabled: ptr(true)})
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

func TestSchedule_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@hourly"})
	path := "/schedules/" + created.str("id")

	// ops sees private but may not run it, and did not make the schedule.
	denied := f.call(t, asOps, http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)

	deleted := f.call(t, asAdmin, http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status)

	again := f.call(t, asAdmin, http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNotFound, again.Status)
}

// An admin picks who a schedule runs as; the scheduler starts its runs
// as that user.
func TestSchedule_RunAs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	res := f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "* * * * *", User: asOps})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, asOps, res.str("user_name"))
	assert.Equal(t, f.users[asOps].ID, res.str("user_id"))

	// By ID, and back to the admin on an update.
	byID := f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily", User: f.users[asLimited].ID})
	assert.Equal(t, asLimited, byID.str("user_name"))
	back := f.call(t, asAdmin, http.MethodPut, "/schedules/"+byID.str("id"), api.ScheduleRequest{User: asAdmin})
	assert.Equal(t, http.StatusOK, back.Status)
	assert.Equal(t, asAdmin, back.str("user_name"))

	// The scheduler fires it as ops.
	loop := scheduler.NewLoop(f.opts.Schedules, f.opts.Users, f.opts.Catalogue, f.opts.Runner)
	loop.Now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	started, err := loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)
	runs, err := f.opts.Runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "plan"})
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, asOps, runs[0].UserName)
	assert.Equal(t, int64(1), runs[0].Scheduled)
}

// Only an admin names another user, who must exist and may run the job.
func TestSchedule_RunAsRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	cases := []struct {
		name   string
		user   string
		req    api.ScheduleRequest
		status int
	}{
		{"not an admin", asOps, api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily", User: asLimited}, http.StatusForbidden},
		{"unknown user", asAdmin, api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily", User: "nobody"}, http.StatusBadRequest},
		{"may not run", asAdmin, api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@daily", User: asOps}, http.StatusBadRequest},
		{"sees no folder", asAdmin, api.ScheduleRequest{Folder: "private", Job: "deploy", Expression: "@daily", User: asLimited}, http.StatusBadRequest},
		{"a viewer", asAdmin, api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily", User: asViewer}, http.StatusBadRequest},
	}
	for _, c := range cases {
		res := f.call(t, c.user, http.MethodPost, "/schedules", c.req)
		assert.Equal(t, c.status, res.Status, c.name)
	}

	// Naming yourself is no pick.
	own := f.call(t, asOps, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: "infra", Job: "plan", Expression: "@daily", User: asOps})
	assert.Equal(t, http.StatusCreated, own.Status)
	refused := f.call(t, asOps, http.MethodPut, "/schedules/"+own.str("id"), api.ScheduleRequest{User: asLimited})
	assert.Equal(t, http.StatusForbidden, refused.Status)
}
