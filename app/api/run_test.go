package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// A run's view carries the title its terminal set.
func TestRunView_title(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.RunView{Run: model.Run{ID: "r1", Title: "deploying"}})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"title":"deploying"`)
}

// A run's view links its report page when it has one, and leaves the
// field out otherwise.
func TestRunView_report(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.RunView{Run: model.Run{ID: "r1"}, Report: "/runs/r1/report"})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"report":"/runs/r1/report"`)
	data, err = json.Marshal(api.RunView{Run: model.Run{ID: "r1"}})
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"report"`)
}

// listRoutes returns the routes over a store of runs, with the folder
// infra and its job build.
func listRoutes(runs *memRuns) http.Handler {
	build := &folder.Job{Folder: "infra", ID: "build"}
	build.Name = "build"
	infra := &folder.Folder{ID: "infra", Dir: ".", Jobs: []*folder.Job{build}}
	return api.New(api.Options{Catalogue: catalogue{"infra": infra}, Runs: runs, Runner: runner.New(runner.Options{})}).Routes()
}

// listed returns the IDs of the runs GET /runs answers for a query.
func listed(t *testing.T, h http.Handler, query string) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/runs"+query, nil)
	req = req.WithContext(access.WithUser(req.Context(), rerunAdmin))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Runs []model.Run `json:"runs"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	ids := []string{}
	for _, run := range body.Runs {
		ids = append(ids, run.ID)
	}
	return ids
}

// The listing takes the parent filter: ?parent= answers the runs one run
// dispatched, ?roots=1 the runs nothing dispatched, and a request naming
// neither answers every run, as it did before either was there.
func TestListRuns_Parent(t *testing.T) {
	t.Parallel()
	h := listRoutes(&memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "build", State: model.StateOK},
		{ID: "r2", Folder: "infra", Job: model.PromptJob, State: model.StateOK, ParentID: "r1", ParentKind: storage.ParentRun},
		{ID: "r3", Folder: "infra", Job: "build", State: model.StateOK, ParentID: "wr1", ParentKind: storage.ParentWorkflowRun},
	}})

	assert.Equal(t, []string{"r3", "r2", "r1"}, listed(t, h, ""))
	assert.Equal(t, []string{"r2"}, listed(t, h, "?parent=r1"))
	assert.Equal(t, []string{"r3"}, listed(t, h, "?parent=wr1"))
	assert.Equal(t, []string{"r1"}, listed(t, h, "?roots=1"))
	assert.Len(t, listed(t, h, "?parent=gone"), 0)
}
