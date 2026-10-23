package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
)

// An admin sets the rules, a size as text; anyone else may not, and a
// wrong age or size answers 400.
func TestRetention_Save(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	body := map[string]map[string]string{"recordings": {"age": "7d", "limit": "500MB"}, "artifacts": {"age": "24h"}}
	res := f.call(t, asAdmin, http.MethodPut, "/retention", body)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	settings, err := f.opts.Cleaner.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, retention.Rule{Age: retention.Week, Limit: 500 << 20}, settings.Recordings)
	assert.Equal(t, retention.Rule{Age: retention.Day}, settings.Artifacts)
	assert.Equal(t, retention.Rule{Age: retention.Never}, settings.SecretReads)

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/retention", body).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/retention", map[string]map[string]string{"recordings": {"age": "3d"}}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/retention", map[string]map[string]string{"recordings": {"limit": "lots"}}).Status)
}

// A run whose recording expired: the cast answers 410 saying so, the
// frame shows it, and a summary answers 410 with the same words.
func TestRetention_ExpiredRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := &model.Run{Folder: "infra", Job: "plan", UserName: "ops", State: model.StateOK, CastPath: "infra/plan/gone.cast", TermCols: 40, TermRows: 5}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))

	cast := f.call(t, asViewer, http.MethodGet, "/runs/"+run.ID+"/cast", nil)
	assert.Equal(t, http.StatusGone, cast.Status)
	assert.Contains(t, cast.str("error"), "expired")
	svg := f.call(t, asViewer, http.MethodGet, "/runs/"+run.ID+"/frame.svg", nil)
	assert.Equal(t, http.StatusOK, svg.Status)
	assert.Contains(t, string(svg.Raw), "expired")
	text := f.call(t, asViewer, http.MethodGet, "/runs/"+run.ID+"/frame.txt?all=1", nil)
	assert.Contains(t, string(text.Raw), "The recording of this run expired")
	summary := f.call(t, asAdmin, http.MethodPost, "/runs/"+run.ID+"/summary", nil)
	assert.Equal(t, http.StatusGone, summary.Status, string(summary.Raw))
}

// @cleanup is the root's, for admins alone.
func TestRetention_CleanupJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: "@cleanup", Cols: 80, Rows: 24})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}
	res := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "@cleanup", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusNotFound, res.Status, "the root's alone")
}
