package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRun_Start(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "plan", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))

	var view api.RunView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.True(t, model.ValidID(view.ID))
	assert.Equal(t, "infra", view.Folder)
	assert.Equal(t, "plan", view.Job)
	assert.Equal(t, "ops", view.UserName)
	assert.True(t, view.Live)
	assert.NotNil(t, view.Status)

	run := f.finished(t, view.ID)
	assert.Equal(t, model.StateOK, run.State)
}

func TestRun_StartRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	cases := []struct {
		user, folder, job string
		status            int
	}{
		{asViewer, "infra", "plan", http.StatusForbidden},
		{asOps, "private", "deploy", http.StatusForbidden},
		{asLimited, "private", "deploy", http.StatusNotFound},
		{asOps, "nowhere", "plan", http.StatusNotFound},
		{asOps, "infra", "nothing", http.StatusNotFound},
	}
	for _, c := range cases {
		res := f.call(t, c.user, http.MethodPost, "/runs", api.RunRequest{Folder: c.folder, Job: c.job})
		assert.Equal(t, c.status, res.Status, c.user+" "+c.folder+"/"+c.job)
	}

	res := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: "private", Job: "deploy"})
	assert.Equal(t, http.StatusCreated, res.Status)
	f.finished(t, res.str("id"))
}

func TestRun_StartBusy(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	first := f.start(t, asOps, "infra", "slow")

	again := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "slow"})
	assert.Equal(t, http.StatusConflict, again.Status)
	assert.Equal(t, "already running", again.str("error"))
	assert.Equal(t, first, again.str("run", "id"))

	forced := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "slow", Force: true})
	assert.Equal(t, http.StatusCreated, forced.Status, string(forced.Raw))
	assert.NotEqual(t, first, forced.str("id"))
	assert.Equal(t, true, forced.get("forced") != nil && forced.get("forced").(float64) == 1)

	killed := f.finished(t, first)
	assert.Equal(t, model.StateKilled, killed.State)
}

func TestRun_StartAll(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	denied := f.call(t, asViewer, http.MethodPost, "/runs/all", api.RunAllRequest{Folder: "infra"})
	assert.Equal(t, http.StatusOK, denied.Status)
	assert.Len(t, denied.list("started"), 0)
	assert.Equal(t, "not allowed", denied.str("skipped", "plan"))

	missing := f.call(t, asOps, http.MethodPost, "/runs/all", api.RunAllRequest{Folder: "nowhere"})
	assert.Equal(t, http.StatusNotFound, missing.Status)

	res := f.call(t, asOps, http.MethodPost, "/runs/all", api.RunAllRequest{Folder: "infra"})
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Len(t, res.list("started"), 5)

	again := f.call(t, asOps, http.MethodPost, "/runs/all", api.RunAllRequest{Folder: "infra"})
	assert.Equal(t, http.StatusOK, again.Status)
	assert.Contains(t, again.str("skipped", "slow"), "running")
}

func TestRun_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	plan := f.start(t, asOps, "infra", "plan")
	deploy := f.start(t, asAdmin, "private", "deploy")
	f.finished(t, plan)
	f.finished(t, deploy)

	all := f.call(t, asViewer, http.MethodGet, "/runs", nil)
	assert.Equal(t, http.StatusOK, all.Status)
	assert.Len(t, all.list("runs"), 2)

	limited := f.call(t, asLimited, http.MethodGet, "/runs", nil)
	assert.Len(t, limited.list("runs"), 1)
	assert.Equal(t, plan, limited.list("runs")[0].(map[string]any)["id"])

	byJob := f.call(t, asOps, http.MethodGet, "/runs?folder=private&job=deploy&limit=5", nil)
	assert.Len(t, byJob.list("runs"), 1)

	none := f.call(t, asOps, http.MethodGet, "/runs?state=running", nil)
	assert.Len(t, none.list("runs"), 0)
}

func TestRun_Get(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "fail")
	f.finished(t, id)

	res := f.call(t, asViewer, http.MethodGet, "/runs/"+id, nil)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Equal(t, model.StateFailed, res.str("state"))
	assert.Equal(t, float64(3), res.get("exit_code"))
	assert.Equal(t, model.StateFailed, res.str("status", "state"))

	hidden := f.call(t, asLimited, http.MethodGet, "/runs/"+f.start(t, asAdmin, "private", "deploy"), nil)
	assert.Equal(t, http.StatusNotFound, hidden.Status)

	missing := f.call(t, asOps, http.MethodGet, "/runs/"+model.NewID(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

func TestRun_StopAndKill(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "slow")
	denied := f.call(t, asViewer, http.MethodPost, "/runs/"+id+"/stop", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)

	stopped := f.call(t, asOps, http.MethodPost, "/runs/"+id+"/stop", nil)
	assert.Equal(t, http.StatusNoContent, stopped.Status)
	assert.Equal(t, model.StateStopped, f.finished(t, id).State)

	id = f.start(t, asAdmin, "infra", "slow")
	killed := f.call(t, asAdmin, http.MethodPost, "/runs/"+id+"/kill", nil)
	assert.Equal(t, http.StatusNoContent, killed.Status)
	assert.Equal(t, model.StateKilled, f.finished(t, id).State)

	gone := f.call(t, asAdmin, http.MethodPost, "/runs/"+model.NewID()+"/kill", nil)
	assert.Equal(t, http.StatusNotFound, gone.Status)
}

func TestRun_Frames(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "plan")
	f.finished(t, id)

	text := f.call(t, asViewer, http.MethodGet, "/runs/"+id+"/frame.txt", nil)
	assert.Equal(t, http.StatusOK, text.Status)
	assert.Contains(t, string(text.Raw), "planned")

	svg := f.call(t, asViewer, http.MethodGet, "/runs/"+id+"/frame.svg", nil)
	assert.Equal(t, http.StatusOK, svg.Status)
	assert.True(t, strings.HasPrefix(string(svg.Raw), "<svg"))
	assert.Contains(t, string(svg.Raw), "planned")

	missing := f.call(t, asViewer, http.MethodGet, "/runs/"+model.NewID()+"/frame.svg", nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

func TestRun_Cast(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "plan")
	f.finished(t, id)

	var cast response
	assert.Eventually(t, func() bool {
		cast = f.call(t, asViewer, http.MethodGet, "/runs/"+id+"/cast", nil)
		return cast.Status == http.StatusOK && strings.Contains(string(cast.Raw), "planned")
	}, 5*time.Second, 20*time.Millisecond)
	assert.Contains(t, string(cast.Raw), `"version":2`)

	bare := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), bare))
	none := f.call(t, asViewer, http.MethodGet, "/runs/"+bare.ID+"/cast", nil)
	assert.Equal(t, http.StatusNotFound, none.Status)
}
