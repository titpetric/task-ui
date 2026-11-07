package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

func TestRun(t *testing.T) {
	t.Parallel()

	run := view.Run{Run: model.Run{Folder: "infra", Job: "plan", UserName: "tit petric", Forced: 1}, Live: true, Attached: 2}
	assert.Equal(t, "TP", run.Initials())
	assert.Equal(t, "infra/plan", run.Key())
	assert.True(t, run.IsForced())
	assert.False(t, run.IsScheduled())
	assert.True(t, run.Live)
	assert.Equal(t, 2, run.Attached)

	run.Scheduled, run.Forced = 1, 0
	assert.True(t, run.IsScheduled())
	assert.False(t, run.IsForced())
}

// A run whose recording expired says so where its output goes.
func TestRun_ExpiredText(t *testing.T) {
	t.Parallel()
	assert.Contains(t, view.Run{Expired: true}.ExpiredText(), "expired under the Data retention settings")
}

// A run that names itself in its terminal shows the name, escaped, on
// the dashboard's tile and control slot, the runs list and its page;
// the run window has the place for it.
func TestRender_runTitle(t *testing.T) {
	t.Parallel()
	v := shared(t)
	for _, name := range []string{"home", "runs", "run"} {
		page := admin(t)
		page.Runs[0].Title = "<b>deploying</b>"
		page.Run.Title = page.Runs[0].Title
		run := page.Running[page.Runs[0].Key()]
		run.Title = page.Runs[0].Title
		page.Running[run.Key()] = run
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, page), name)
		html := buf.String()
		assert.Contains(t, html, `class="term-title"`, name)
		assert.Contains(t, html, "&lt;b&gt;deploying&lt;/b&gt;", name)
		assert.NotContains(t, html, "<b>deploying", name)
		assert.Contains(t, html, `id="run-subtitle"`, name)
	}
}

// A finished run's page replays its recording and reads it into Output,
// and links no .cast to download.
func TestRender_runRecording(t *testing.T) {
	t.Parallel()
	v := shared(t)
	page := admin(t)
	page.Run.State, page.Run.CastPath = model.StateOK, "infra/plan/x.cast"
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "run", page))
	html := buf.String()
	assert.Contains(t, html, `data-replay="`+page.Run.ID+`"`)
	assert.Contains(t, html, `data-output="`+page.Run.ID+`"`)
	assert.NotContains(t, html, "/cast")
	assert.NotContains(t, html, `.cast"`)
	assert.NotContains(t, html, ">Recording</a>")
}

// A run that ended has Run again in its head, with the arguments it ran
// with and no args line; a prompt's run and the Notes agent's have none,
// because POST /api/runs refuses their jobs.
func TestRender_runAgain(t *testing.T) {
	t.Parallel()
	v := shared(t)
	render := func(page *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "run", page))
		return buf.String()
	}
	// The button posts the run's own ID: the endpoint reads the inputs off
	// the row, so no folder, job or argument list is in the markup. The
	// layout holds the window's own Run again, hidden, under the same
	// label, which is why the head's button is read by its attribute.
	page := admin(t)
	starts := `data-rerun="` + page.Run.ID + `"`
	assert.NotContains(t, render(page), starts, "still running")

	page.Run.State = model.StateKilled
	html := render(page)
	assert.Contains(t, html, starts+` data-title="infra / plan">Run again<`)
	assert.NotContains(t, html, "data-run-args=")
	assert.NotContains(t, html, `data-job="plan" data-title="infra / plan">Run again<`)

	// A prompt's run and the Notes agent's repeat the question, which
	// takes prompt/edit and not the job's gate.
	page.Run.Job = model.PromptJob
	assert.Contains(t, render(page), starts, "a prompt's run")

	page.Run.Job = model.NoteJob
	assert.Contains(t, render(page), starts, "the Notes agent's run")
}

// A run's page lists the runs it dispatched, each a link with its job
// and how it ended, and a run that dispatched none lists nothing.
func TestRender_runChildren(t *testing.T) {
	t.Parallel()
	v := shared(t)
	page := admin(t)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "run", page))
	assert.NotContains(t, buf.String(), "Runs it started")

	page.Children = []model.Run{
		{ID: "01child", Folder: "infra", Job: model.PromptJob, State: model.StateOK},
		{ID: "02child", Folder: "infra", Job: "plan", State: model.StateFailed, ExitCode: 2},
	}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "run", page))
	html := buf.String()
	assert.Contains(t, html, "<h3>Runs it started <small>oldest first</small></h3>")
	assert.Contains(t, html, `<a href="/runs/01child">infra / @claude-prompt</a>`)
	assert.Contains(t, html, `<span class="state" data-run-id="02child" data-finished="true"><i class="dot failed"></i>failed <span class="faint">(2)</span></span>`)
}

// A recent run that dispatched others draws the count of them, and one
// that dispatched none draws no badge.
func TestRender_recentChildren(t *testing.T) {
	t.Parallel()
	v := shared(t)
	page := admin(t)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", page))
	assert.NotContains(t, buf.String(), `<span class="badge" title="Runs it started">`)

	page.Recents[0].Entries[0].Children = 3
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "home", page))
	assert.Contains(t, buf.String(), `<span class="badge" title="Runs it started">3</span>`)
}

// A run that ended has Summarize in its head for a user who may ask,
// and none once the summary is there to read as a Report. A summary
// Claude is still writing says so where the button stood: the button
// would post the question a second time.
func TestRender_runSummarize(t *testing.T) {
	t.Parallel()
	v := shared(t)
	render := func(page *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "run", page))
		return buf.String()
	}
	page := admin(t)
	page.Run.State = model.StateFailed
	assert.NotContains(t, render(page), "data-summarize-run", "no Summaries")

	page.Summaries = true
	html := render(page)
	assert.Contains(t, html, `data-summarize-run="`+page.Run.ID+`"`)
	assert.Contains(t, html, `data-question="Summarize this run."`)
	assert.NotContains(t, html, ">summarizing<", "nothing asked yet")

	page.RunReports = map[string]*note.Report{page.Run.ID: {RunID: page.Run.ID, Pending: true}}
	html = render(page)
	assert.NotContains(t, html, "data-summarize-run", "a summary under way")
	assert.Contains(t, html, `<span class="faint run-denied">summarizing</span>`)

	page.RunReports = map[string]*note.Report{page.Run.ID: {RunID: page.Run.ID, Body: "## Passed"}}
	page.Report = &note.Report{RunID: page.Run.ID}
	html = render(page)
	assert.NotContains(t, html, "data-summarize-run", "a summary to read")
	assert.NotContains(t, html, ">summarizing<", "a summary to read")
	assert.Contains(t, html, `<a class="btn" href="/runs/`+page.Run.ID+`/report">Report</a>`)

	page.RunReports, page.Report, page.Run.State = nil, nil, model.StateRunning
	assert.NotContains(t, render(page), "data-summarize-run", "still running")
}
