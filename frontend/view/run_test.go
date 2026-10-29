package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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
	v, err := view.New()
	assert.NoError(t, err)
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
	v, err := view.New()
	assert.NoError(t, err)
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
