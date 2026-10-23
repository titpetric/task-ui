package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/retention"
)

// recordedRun adds a finished run with a recording of a size, or none
// when gone, as the retention rules leave it.
func recordedRun(t *testing.T, f *fixture, size int, gone bool) *model.Run {
	t.Helper()
	run := f.run(t, "infra", "plan", asOps, model.StateOK)
	path, err := f.opts.Recordings.Path(run)
	assert.NoError(t, err)
	run.CastPath = f.opts.Recordings.Relative(path)
	run.TermCols, run.TermRows = 80, 24
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))
	if !gone {
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, make([]byte, size), 0o644))
	}
	return run
}

// The page shows what each kind holds and its rule; an admin edits the
// rules and cleans up now, a user who may see the settings reads them,
// and a viewer may not open the page.
func TestDataRetention(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	recordedRun(t, f, 2048, false)
	recordedRun(t, f, 2048, false)
	assert.NoError(t, f.opts.Cleaner.Save(context.Background(), retention.Settings{Recordings: retention.Rule{Age: retention.Week, Limit: 1 << 20}, Artifacts: retention.Rule{Age: retention.Never}, SecretReads: retention.Rule{Age: retention.Never}}))

	body := f.page(t, asAdmin, "/settings/retention")
	assert.Contains(t, body, "<h2>Data retention</h2>")
	assert.Contains(t, body, "<h3>Run recordings</h3>")
	assert.Contains(t, body, "<h3>Artifact history</h3>")
	assert.Contains(t, body, "<h3>Secret read log</h3>")
	assert.Contains(t, body, "2 runs, 4.0 KB")
	assert.Contains(t, body, `data-average="2048"`)
	assert.Contains(t, body, `name="recordings.age" value="7d" checked`)
	assert.Contains(t, body, `name="artifacts.age" value="never" checked`)
	assert.Contains(t, body, `value="1.0 MB"`)
	assert.Contains(t, body, "about 512 runs")
	assert.Contains(t, body, `data-api="/api/retention"`)
	assert.Contains(t, body, `data-job="@cleanup"`)
	assert.Contains(t, body, `rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: Instance</a>`)
	assert.NotContains(t, body, "<fieldset class=\"field\" disabled>")

	assert.NoError(t, f.opts.Users.Policies().Put(context.Background(), permission.KindUser, asOps, "secret/view", permission.Allow))
	read := f.page(t, asOps, "/settings/retention")
	assert.Contains(t, read, "<fieldset class=\"field\" disabled>")
	assert.Contains(t, read, "An admin changes these rules.")
	assert.NotContains(t, read, `data-job="@cleanup"`)
	assert.NotContains(t, read, `type="submit">Save`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/settings/retention").Status)
}

// A run whose recording expired says so on its page, with no Replay and
// no Recording download.
func TestRun_ExpiredRecording(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := recordedRun(t, f, 0, true)
	body := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, body, "The run has finished; its recording expired.")
	assert.Contains(t, body, "The recording of this run expired under the Data retention settings")
	assert.NotContains(t, body, `data-replay="`+run.ID+`"`)
	assert.NotContains(t, body, `/api/runs/`+run.ID+`/cast" download`)
	assert.NotContains(t, body, `data-output="`+run.ID+`"`)
}
