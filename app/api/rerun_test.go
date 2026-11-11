package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
)

// rerunCall posts a rerun of a run's ID as a user.
func rerunCall(t *testing.T, h http.Handler, user *model.User, id string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/runs/"+id+"/rerun", nil)
	req = req.WithContext(access.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

// rerunRoutes returns the routes over the folder infra, its job plan,
// which takes args, and build, which takes none, with the hours of the
// workdays and the clock on a Saturday: a rerun that reaches the runner
// is refused there, and says so.
func rerunRoutes(runs *memRuns) http.Handler {
	plan := &folder.Job{Folder: "infra", ID: "plan"}
	plan.Name, plan.TakesArgs = "plan", true
	build := &folder.Job{Folder: "infra", ID: "build"}
	build.Name = "build"
	infra := &folder.Folder{ID: "infra", Dir: ".", Hours: "mon-fri 9-17", Jobs: []*folder.Job{plan, build}}
	saturday := func() time.Time { return time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC) }
	return api.New(api.Options{Catalogue: catalogue{"infra": infra}, Runs: runs, Runner: runner.New(runner.Options{Now: saturday})}).Routes()
}

// rerunAdmin and rerunViewer are the callers of the rerun tests: one who
// may run the jobs of every folder, one who may see them alone.
var (
	rerunAdmin  = &model.User{ID: "u1", Name: "black", Role: model.RoleAdmin}
	rerunViewer = &model.User{ID: "u2", Name: "ops", Role: model.RoleViewer}
)

// The request a stored run reruns with: its folder, the directory inside
// it, its job and the arguments it passed, the job of a run on a host
// naming the host; it forces nothing and takes no summary.
func TestRerunRequest(t *testing.T) {
	t.Parallel()
	req := api.RerunRequest(&model.Run{ID: "r1", Folder: "infra", Dir: "deploy", Job: "plan", Args: model.EncodeList([]string{"-x", "web"}), UserName: "ops"})
	assert.Equal(t, "infra", req.Folder)
	assert.Equal(t, "deploy", req.Dir)
	assert.Equal(t, "plan", req.Job)
	assert.Equal(t, api.RunArgs{"-x", "web"}, req.Args)
	assert.False(t, req.Force)
	assert.False(t, req.Summarize)
	assert.Equal(t, "", req.Reason)

	// A run on a host keeps its host in its job and the command in its args.
	host := api.RerunRequest(&model.Run{ID: "r2", Folder: model.Default, Job: model.HostJob("dev1"), Target: "dev1", Args: model.EncodeList([]string{"uptime"})})
	assert.Equal(t, model.HostJob("dev1"), host.Job)
	assert.Equal(t, api.RunArgs{"uptime"}, host.Args)

	bare := api.RerunRequest(&model.Run{ID: "r3", Folder: "infra", Job: "build"})
	assert.Len(t, bare.Args, 0)
}

// A run that ended runs again from its ID alone, on the start path of
// POST /api/runs, with the folder, job and arguments of the stored run:
// the folder's hours refuse the start here and name themselves, and
// arguments of a job that takes none are refused as they are there.
func TestRerun(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK, UserID: rerunViewer.ID, UserName: rerunViewer.Name, Args: model.EncodeList([]string{"-x"})},
		{ID: "r2", Folder: "infra", Job: "build", State: model.StateKilled, Args: model.EncodeList([]string{"-x"})},
	}})

	code, out := rerunCall(t, h, rerunAdmin, "r1")
	assert.Equal(t, http.StatusForbidden, code, out)
	assert.Equal(t, "out of hours", out["error"])
	assert.Equal(t, true, out["out_of_hours"])
	assert.Equal(t, "mon-fri 9-17", out["hours"])

	code, out = rerunCall(t, h, rerunAdmin, "r2")
	assert.Equal(t, http.StatusBadRequest, code, out)
	assert.Equal(t, runner.ErrNoArgs.Error(), out["error"])
}

// An ID of no run, a run of a folder the caller may not see and a run of
// a folder that is gone are each 404.
func TestRerun_NotFound(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK},
		{ID: "r2", Folder: "gozo", Job: "plan", State: model.StateOK},
	}})

	code, _ := rerunCall(t, h, rerunAdmin, "nope")
	assert.Equal(t, http.StatusNotFound, code)

	// The folder is not in the catalogue the start resolves through.
	code, out := rerunCall(t, h, rerunAdmin, "r2")
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "no such folder", out["error"])

	// The caller's own folders leave the run's out.
	elsewhere := &model.User{ID: "u3", Name: "dana", Role: model.RoleAdmin, Folders: model.EncodeList([]string{"malta"})}
	code, _ = rerunCall(t, h, elsewhere, "r1")
	assert.Equal(t, http.StatusNotFound, code)
}

// A run that has not ended is refused: it is going, or it waits.
func TestRerun_Going(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "plan", State: model.StateRunning},
		{ID: "r2", Folder: "infra", Job: "plan", State: model.StateQueued},
	}})

	code, out := rerunCall(t, h, rerunAdmin, "r1")
	assert.Equal(t, http.StatusConflict, code, out)
	assert.Equal(t, "you can view this run, but it hasn't ended yet", out["error"])

	code, _ = rerunCall(t, h, rerunAdmin, "r2")
	assert.Equal(t, http.StatusConflict, code)
}

// The caller's own permission decides the rerun, never the permission of
// whoever started the run: a user who may not run the job, or may not
// enter a shell, is refused what POST /api/runs refuses them.
func TestRerun_Permission(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK, UserID: rerunAdmin.ID, UserName: rerunAdmin.Name},
		{ID: "r2", Folder: "infra", Job: folder.ShellID, State: model.StateOK, UserID: rerunAdmin.ID, UserName: rerunAdmin.Name},
	}})

	code, out := rerunCall(t, h, rerunViewer, "r1")
	assert.Equal(t, http.StatusForbidden, code, out)
	assert.Equal(t, "you can view this job, but you don't have permission to run it", out["error"])

	code, out = rerunCall(t, h, rerunViewer, "r2")
	assert.Equal(t, http.StatusForbidden, code, out)
	assert.Equal(t, "an interactive shell or Claude terminal needs the shell permission", out["error"])
}

// A prompt's run and the Notes agent's take the board's path and not the
// job's: POST /api/runs refuses the job either of them runs under, so a
// rerun that took it would refuse every run of the Orb.
//
// With no board there is nothing to ask, which is 404 and not the job's
// 400. The question this answers is covered where a board is real.
func TestRerun_Claude(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: model.PromptJob, State: model.StateOK, UserID: rerunAdmin.ID},
		{ID: "r2", Folder: model.Default, Job: model.NoteJob, State: model.StateOK, UserID: rerunAdmin.ID},
	}})

	for _, id := range []string{"r1", "r2"} {
		code, out := rerunCall(t, h, rerunAdmin, id)
		assert.Equal(t, http.StatusNotFound, code, out)
		assert.NotContains(t, out["error"], "POST /api/prompts")
		assert.NotContains(t, out["error"], "Notes board")
	}
}

// The board's path takes what POST /api/prompts takes, asked through a
// rerun as well: a user who sees their own prompt run and may not ask
// again is refused the permission, before any board is reached.
func TestRerun_Asking(t *testing.T) {
	t.Parallel()
	h := rerunRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: model.PromptJob, State: model.StateOK, UserID: rerunViewer.ID, UserName: rerunViewer.Name},
	}})

	code, out := rerunCall(t, h, rerunViewer, "r1")
	assert.Equal(t, http.StatusForbidden, code, out)
	assert.Equal(t, "you need the prompt/edit permission", out["error"])
}
