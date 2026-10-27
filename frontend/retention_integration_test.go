//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// The page is one table with a row per area in order, the fixed rows
// among them and no key read log; the selects show the saved rule, a
// limit the list lacks as an option of its own; an admin edits the
// rules and cleans up now, a user who may see the settings reads them,
// and a viewer may not open the page.
func TestDataRetention(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	recordedRun(t, f, 2048, false)
	recordedRun(t, f, 2048, false)
	settings := retention.Defaults()
	settings[retention.Recordings] = retention.Rule{Age: retention.Week, Limit: 1 << 20}
	settings[retention.Runs] = retention.Rule{Age: retention.Year, Limit: 3 << 30}
	settings[retention.Artifacts] = retention.Rule{Age: retention.Never}
	assert.NoError(t, f.opts.Cleaner.Save(context.Background(), settings))
	body := f.page(t, asAdmin, "/settings/retention")
	assert.Contains(t, body, `Applied by <a href="/runs?folder=internal&amp;job=@cleanup" title="The runs of @cleanup">@cleanup</a>, which <span class="faint">has not run yet</span>.`, "the lead names the job before its first run")
	cleanup := f.run(t, model.Default, "@cleanup", asAdmin, model.StateOK)

	body = f.page(t, asAdmin, "/settings/retention")
	assert.Contains(t, body, "<h2>Data retention</h2>")
	assert.Contains(t, body, "<thead><tr><th>Area</th><th>What goes</th><th>Keep</th><th>Size limit</th><th>Holds now</th></tr></thead>")
	assert.NotContains(t, body, "Last cleanup")
	assert.Contains(t, body, `Applied by <a href="/runs?folder=internal&amp;job=@cleanup" title="The runs of @cleanup">@cleanup</a>, which last ran <a href="/runs/`+cleanup.ID+`"`)
	last := -1
	for _, area := range retention.Areas {
		at := strings.Index(body, `<td class="key">`+area.Title+`</td>`)
		assert.Greater(t, at, last, area.Title)
		last = at
	}
	assert.NotContains(t, body, "Key read log")
	assert.Contains(t, body, `<select name="recordings.age" aria-label="Keep Run recordings for">`)
	assert.Contains(t, body, `<option value="7d" selected>7 days</option>`)
	assert.Contains(t, body, `<option value="365d" selected>365 days</option>`)
	assert.Contains(t, body, `<option value="1048576" selected>1 MB</option>`, "a limit the list lacks shows as set")
	assert.Contains(t, body, `<option value="3221225472" selected>3 GB</option>`)
	assert.Contains(t, body, `<option value="never" selected>Never</option>`)
	assert.Contains(t, body, `<select name="prompts.age"`)
	assert.NotContains(t, body, `name="prompts.limit"`, "no size on a row without one")
	assert.Contains(t, body, `<td colspan="2" class="dim">Removed every cleanup, no setting.</td>`)
	assert.Contains(t, body, `<select name="tooling.age" aria-label="Keep Tool history for"`)
	assert.NotContains(t, body, `name="tooling.limit"`, "tool history takes an age, no size limit")
	assert.Contains(t, body, "0 databases, 0 B")
	assert.Contains(t, body, "2 recordings, 4.0 KB")
	assert.Contains(t, body, "3 runs, 4.0 KB", "run history counts the cleanup's run, its size its recordings'")
	assert.Contains(t, body, `<a href="/runs/`+cleanup.ID+`"`)
	assert.Contains(t, body, `<i class="dot ok"></i>ok`)
	assert.Contains(t, body, `data-api="/api/retention"`)
	assert.Contains(t, body, `data-job="@cleanup"`)
	assert.Contains(t, body, `rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: Instance</a>`)
	assert.NotContains(t, body, " disabled>")

	assert.NoError(t, f.opts.Users.Policies().Put(context.Background(), permission.KindUser, asOps, "secret/view", permission.Allow))
	read := f.page(t, asOps, "/settings/retention")
	assert.Contains(t, read, `<select name="recordings.age" aria-label="Keep Run recordings for" disabled>`)
	assert.Contains(t, read, "An admin changes the rules.")
	assert.NotContains(t, read, `data-job="@cleanup"`)
	assert.NotContains(t, read, `type="submit">Save`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/settings/retention").Status)
}

// A run whose recording expired says so on its page, with no Replay.
func TestRun_ExpiredRecording(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := recordedRun(t, f, 0, true)
	body := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, body, "The run has finished; its recording expired.")
	assert.Contains(t, body, "The recording of this run expired under the Data retention settings")
	assert.NotContains(t, body, `data-replay="`+run.ID+`"`)
	assert.NotContains(t, body, `data-output="`+run.ID+`"`)
}
