package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	assert.True(t, view.CanRun)
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

// The API has no run all route: a workflow chains a folder's jobs.
func TestRun_NoStartAll(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asOps, http.MethodPost, "/runs/all", map[string]string{"folder": "infra"})
	assert.Equal(t, http.StatusMethodNotAllowed, res.Status)
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
	assert.Equal(t, false, res.get("can_run"))
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
	// Trimmed to what there is; the full screen on request.
	assert.Contains(t, string(svg.Raw), `viewBox="0 0 672.0 136.0"`)
	full := f.call(t, asViewer, http.MethodGet, "/runs/"+id+"/frame.svg?full=1", nil)
	assert.Contains(t, string(full.Raw), `viewBox="0 0 672.0 408.0"`)

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

// Outside the job's hours the API refuses a run; forced with a reason it runs
// and the run keeps the reason.
func TestRun_OutOfHours(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.opts.Catalogue.Find("infra").Hours = "sun 1-2"

	denied := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "plan"})
	assert.Equal(t, http.StatusForbidden, denied.Status, string(denied.Raw))
	assert.Equal(t, "out of hours", denied.str("error"))
	assert.Equal(t, true, denied.get("out_of_hours"))
	assert.Equal(t, "sun 1-2", denied.str("hours"))

	silent := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "plan", Force: true})
	assert.Equal(t, http.StatusBadRequest, silent.Status)
	assert.Equal(t, true, silent.get("out_of_hours"))

	forced := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "plan", Force: true, Reason: "customer is waiting"})
	assert.Equal(t, http.StatusCreated, forced.Status, string(forced.Raw))
	assert.Equal(t, "customer is waiting", forced.str("reason"))
	assert.Equal(t, model.StateOK, f.finished(t, forced.str("id")).State)

}

// A run whose session has ended still has a last screen: the one its
// recording ends on.
func TestFrame_FromRecording(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := model.NewID()
	rel := filepath.Join("infra", "plan", id+".cast")
	path := filepath.Join(f.opts.Recordings.Root, rel)
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte("{\"version\":2,\"width\":40,\"height\":5}\n[0.1,\"o\",\"planned, from the recording\"]\n"), 0o644))
	run := &model.Run{ID: id, Folder: "infra", Job: "plan", State: model.StateOK, UserName: "ops", CastPath: rel, TermCols: 40, TermRows: 5}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))

	res := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Contains(t, string(res.Raw), "planned, from the recording")
	svg := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.svg", nil)
	assert.Equal(t, http.StatusOK, svg.Status)
	assert.Contains(t, svg.Headers.Get("Content-Type"), "image/svg+xml")
	assert.Equal(t, "true", svg.Headers.Get("X-Output-Complete")) // one line: nothing to expand

	// ?all=1 is the whole output, where the last screen of a 5 row
	// terminal holds only the last lines.
	long := "{\"version\":2,\"width\":40,\"height\":5}\n"
	for i := 0; i < 20; i++ {
		long += "[0.1,\"o\",\"output line " + strconv.Itoa(i) + "\\r\\n\"]\n"
	}
	assert.NoError(t, os.WriteFile(path, []byte(long), 0o644))
	screen := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt", nil)
	assert.NotContains(t, string(screen.Raw), "output line 0\n")
	scrolled := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.svg", nil)
	assert.Equal(t, "", scrolled.Headers.Get("X-Output-Complete"))
	assert.Equal(t, "20", scrolled.Headers.Get("X-Output-Lines"))
	all := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt?all=1", nil)
	assert.Equal(t, http.StatusOK, all.Status)
	assert.Contains(t, string(all.Raw), "output line 0\n")
	assert.Contains(t, string(all.Raw), "output line 19")

	// No session and no recording: nothing to show.
	bare := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK, UserName: "ops"}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), bare))
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/runs/"+bare.ID+"/frame.txt", nil).Status)
}
