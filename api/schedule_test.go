package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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
