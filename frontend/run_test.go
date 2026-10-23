package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRuns(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	plan := f.run(t, "infra", "plan", "ops", model.StateOK)
	deploy := f.run(t, "private", "deploy", "admin", model.StateFailed)

	body := f.page(t, asAdmin, "/runs")
	assert.Contains(t, body, "<h2>Runs</h2>")
	assert.Contains(t, body, `href="/runs/`+plan.ID+`"`)
	assert.Contains(t, body, `href="/runs/`+deploy.ID+`"`)
	// The table follows the runs topic; the page draws on it and never reloads.
	assert.Contains(t, body, `<section class="block" data-sse="runs">`)
	assert.Contains(t, body, `<tr data-run-id="`+plan.ID+`">`)
	// Only the View button opens the window: the row is not a Run button,
	// the folder and the job are plain links to their pages.
	assert.NotContains(t, body, `<tr data-run=`)
	assert.Contains(t, body, `<td data-label="Folder"><a href="/f/_/infra">infra</a></td>`)
	assert.Contains(t, body, `<td class="key"><a href="/f/_/infra/plan">plan</a></td>`)
	assert.NotContains(t, body, `data-live`)
	// A run without a recording has no replay.
	assert.NotContains(t, body, `data-replay="`+plan.ID+`"`)

	// The sidebar lists every recent run; the table only the filtered ones.
	filtered := f.page(t, asAdmin, "/runs?folder=infra&state=ok")
	assert.Contains(t, filtered, `class="btn btn-sm" href="/runs/`+plan.ID+`"`)
	assert.NotContains(t, filtered, `class="btn btn-sm" href="/runs/`+deploy.ID+`"`)
	assert.Contains(t, filtered, `<option value="infra" selected>`)
	assert.Contains(t, filtered, `<option value="ok" selected>`)

	// Summary shows once the Claude integration has a token for the user.
	assert.NotContains(t, body, `data-summary=`)
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-1", "who": "admin"}))
	summed := f.page(t, asAdmin, "/runs")
	assert.Contains(t, summed, `<tbody data-runs-table data-summarize>`)
	assert.Contains(t, summed, `data-summary="/api/runs/`+deploy.ID+`/summary" data-summary-page="/runs/`+deploy.ID+`" data-summary-title="private / deploy"`)
	assert.NotContains(t, f.page(t, asOps, "/runs"), `data-summary=`)

	limited := f.page(t, asLimited, "/runs")
	assert.NotContains(t, limited, deploy.ID)
}

func TestRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "plan", "ops", model.StateFailed)
	run.ExitCode, run.CastPath = 2, "infra/plan/x.cast"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))
	assert.NoError(t, f.opts.Artifacts.Create(context.Background(), &model.Artifact{RunID: run.ID, Folder: "infra", Job: "plan", Name: "report.txt", Path: "x/report.txt", ByteSize: 2048, ContentType: "text/plain"}))

	body := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, body, `<h2>infra / plan <span class="badge bad">failed 2</span></h2>`)
	assert.Contains(t, body, `data-attach="`+run.ID+`">View</button>`)
	assert.Contains(t, body, `href="/api/runs/`+run.ID+`/cast"`)
	assert.Contains(t, body, `<button type="button" class="btn" data-replay="`+run.ID+`" data-replay-title="infra / plan" data-replay-who="ops" title="Replay the recording"`)
	assert.Contains(t, body, "report.txt")
	assert.Contains(t, body, "2.0 KB")
	assert.Contains(t, body, `<span class="who-ran"><span class="user-ref"><span class="avatar sm role-runner" title="ops">OP</span> <span class="user-name">ops</span><details class="more-menu user-more">`)
	// The viewer's menu of ops: the runs by ops, no profile and no
	// permissions without user/view and user/edit.
	assert.Contains(t, body, `href="/runs?user=ops"><svg class="ic"><use href="#i-list"/></svg>Runs by this user</a>`)
	assert.NotContains(t, body, `href="/users/ops"`)
	assert.NotContains(t, body, `href="/users/ops#permissions"`)
	assert.NotContains(t, body, "data-frame")
	assert.NotContains(t, body, "data-delete")

	// The row of a run with a recording replays it beside View.
	assert.Contains(t, f.page(t, asViewer, "/runs"), `data-replay="`+run.ID+`"`)

	admin := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, admin, `data-delete="/api/artifacts/`)
}

func TestRun_Live(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/runs/"+session.ID())
	assert.Contains(t, body, `data-attach="`+session.ID()+`">Attach</button>`)
	assert.Contains(t, body, `data-stop="`+session.ID()+`"`)
	assert.Contains(t, body, `data-kill="`+session.ID()+`"`)
	assert.Contains(t, body, `data-frame="`+session.ID()+`"`)
	assert.Contains(t, body, "nobody attached")

	viewer := f.page(t, asViewer, "/runs/"+session.ID())
	assert.NotContains(t, viewer, "data-stop")
	// A run in progress has nothing to replay yet.
	assert.NotContains(t, viewer, "data-replay")
	assert.Contains(t, viewer, "data-frame")
}

func TestRun_Missing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	hidden := f.run(t, "private", "deploy", "admin", model.StateOK)

	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/runs/"+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/runs/"+hidden.ID).Status)
}

func TestRun_Reason(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "plan", "ops", model.StateOK)
	run.Forced, run.Reason = 1, "customer waiting on the fix"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))

	page := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, page, "<th>Reason</th><td>customer waiting on the fix")
	assert.Contains(t, page, `title="customer waiting on the fix">out of hours</span>`)

	list := f.page(t, asViewer, "/runs")
	assert.Contains(t, list, `title="customer waiting on the fix">out of hours</span>`)
	assert.NotContains(t, list, ">forced</span>")
	assert.Contains(t, list, `<table class="table cards">`)
	assert.Contains(t, list, `data-label="Result"`)
}

// A run someone watches says so in its result; the recents open its page
// and mark by permission what it allows there.
func TestRuns_Watching(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")
	sub := session.Attach(true)
	defer session.Detach(sub)

	ops := f.page(t, asOps, "/runs")
	assert.Contains(t, ops, `running</span> <span class="faint">1 watching</span>`)
	assert.NotContains(t, ops, "<th>Watching</th>")
	assert.Contains(t, ops, `title="Open infra / slow, ops, running"`)
	assert.Contains(t, ops, `<svg class="ic act"><use href="#i-play"/></svg>`)

	viewer := f.page(t, asViewer, "/runs")
	assert.Contains(t, viewer, `title="Open infra / slow, ops, running"`)
	assert.Contains(t, viewer, `<svg class="ic act"><use href="#i-eye"/></svg>`)
}

// The runs come a page at a time, 15 unless ?n= picks 50 or 100, with
// Previous and Next under the table that keep the filters.
func TestRuns_Pages(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	base := time.Now().Add(-time.Hour)
	var infra []string // newest first
	for i := range 40 {
		started := base.Add(time.Duration(i) * time.Second)
		folder := "infra"
		if i%2 == 1 {
			folder = "private"
		}
		run := &model.Run{ID: model.NewIDAt(started), Folder: folder, Job: "plan", UserName: "ops", State: model.StateOK, StartedAt: &started}
		assert.NoError(t, f.opts.Runs.Create(context.Background(), run))
		if folder == "infra" {
			infra = append([]string{run.ID}, infra...)
		}
	}
	rows := func(body string) int { return strings.Count(body, `<tr data-run-id="`) }

	first := f.page(t, asAdmin, "/runs")
	assert.Equal(t, 15, rows(first))
	assert.Contains(t, first, `<select name="n">`)
	assert.Contains(t, first, `<option value="15" selected>15</option>`)
	assert.Contains(t, first, `<nav class="pager" aria-label="Pages">`)
	assert.NotContains(t, first, `rel="prev"`)
	assert.Equal(t, 15, rows(f.page(t, asAdmin, "/runs?n=7")))
	assert.Equal(t, 40, rows(f.page(t, asAdmin, "/runs?n=100")))
	assert.NotContains(t, f.page(t, asAdmin, "/runs?n=100"), `class="pager"`)

	// A limited user's page is full, of their folder only.
	limited := f.page(t, asLimited, "/runs?folder=infra&state=ok&n=50")
	assert.Equal(t, 20, rows(limited))
	limited = f.page(t, asLimited, "/runs")
	assert.Equal(t, 15, rows(limited))
	for _, id := range infra[:15] {
		assert.Contains(t, limited, `<tr data-run-id="`+id+`">`)
	}
	assert.Contains(t, limited, `href="/runs?before=`+infra[14]+`&amp;n=15" rel="next">Next</a>`)

	// Next, then Previous, come back to the same rows; the links carry
	// the filters and the page size.
	next := f.page(t, asLimited, "/runs?folder=infra&state=ok&n=15&before="+infra[14])
	assert.Equal(t, 5, rows(next))
	assert.Contains(t, next, `<tr data-run-id="`+infra[15]+`">`)
	assert.Contains(t, next, `href="/runs?after=`+infra[15]+`&amp;folder=infra&amp;n=15&amp;state=ok" rel="prev">Previous</a>`)
	assert.NotContains(t, next, `rel="next"`)
	back := f.page(t, asLimited, "/runs?folder=infra&state=ok&n=15&after="+infra[15])
	assert.Equal(t, 15, rows(back))
	assert.Contains(t, back, `<tr data-run-id="`+infra[0]+`">`)
	assert.Contains(t, back, `<tr data-run-id="`+infra[14]+`">`)
	assert.NotContains(t, back, `rel="prev"`)
	assert.Contains(t, back, `rel="next"`)
}

// A run of a directory's job names the directory, links it and the job's
// page under it, and lists the arguments it passed, one each.
func TestRun_Dir(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	run := f.run(t, "infra", "custom", "ops", model.StateOK)
	run.Dir, run.Args = "images", model.EncodeList([]string{"ps", "; id"})
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))

	body := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, body, `<h2>infra/images / custom <span`)
	assert.Contains(t, body, `<th>Directory</th><td><a href="/f/_/infra/images">`)
	assert.Contains(t, body, `<th>Job</th><td><a href="/f/_/infra/images/custom">custom</a></td>`)
	assert.Contains(t, body, `<code>ps</code> <code>; id</code>`)

	plain := f.page(t, asViewer, "/runs/"+f.run(t, "infra", "plan", "ops", model.StateOK).ID)
	assert.NotContains(t, plain, "<th>Directory</th>")
	assert.NotContains(t, plain, "<th>Arguments</th>")
}

// A finished run's page holds a prompt about the run, under its output,
// for a user with prompt/edit and a Claude token, and lists the threads
// about the run: read-only with prompt/view. A prompt's own run, in the
// internal folder, has none.
func TestRun_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	const box = `<input type="hidden" name="run" value="`
	run := f.run(t, "infra", "plan", "ops", model.StateFailed)
	assert.NotContains(t, f.page(t, asAdmin, "/runs/"+run.ID), box, "no Claude token, no prompt")

	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	admin := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, admin, `<details class="promptbox" open>`)
	assert.Contains(t, admin, box+run.ID+`">`)
	assert.Contains(t, admin, "ask Claude Code about this run and its output, in infra/")
	assert.Contains(t, admin, `data-notes data-dir="infra"`)
	assert.NotContains(t, f.page(t, asViewer, "/runs/"+run.ID), "promptbox")

	// The threads about the run list on its page alone.
	thread := &model.Note{UserName: "admin", Body: "why did it fail", State: model.NoteDone, WorkspaceID: "ws", Dir: "infra", AboutRunID: run.ID}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, thread))
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, &model.Note{UserName: "admin", Body: "what builds here", WorkspaceID: "ws", Dir: "infra"}))
	admin = f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, admin, `id="note-`+thread.ID+`"`)
	assert.NotContains(t, admin, "what builds here")
	assert.NotContains(t, f.page(t, asAdmin, "/f/_/infra"), "why did it fail")

	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asViewer, "prompt/view", "allow"))
	viewer := f.page(t, asViewer, "/runs/"+run.ID)
	assert.Contains(t, viewer, `id="note-`+thread.ID+`"`)
	assert.NotContains(t, viewer, box)

	// A running run's box follows its threads, hidden while it has none:
	// the summary of Run and summarize comes when it ends.
	running := f.run(t, "infra", "slow", "ops", model.StateRunning)
	live := f.page(t, asAdmin, "/runs/"+running.ID)
	assert.Contains(t, live, `<details class="promptbox" open hidden>`)
	assert.NotContains(t, live, box)

	prompt := f.run(t, model.Internal, "claude-prompt", "admin", model.StateOK)
	assert.NotContains(t, f.page(t, asAdmin, "/runs/"+prompt.ID), "promptbox", "an older prompt's run renders, with no box")

	// A prompt's run is its place's, @claude-prompt, with no box; the
	// runs list shows it to those with prompt/view.
	asked := f.run(t, "infra", model.PromptJob, "admin", model.StateOK)
	page := f.page(t, asAdmin, "/runs/"+asked.ID)
	assert.Contains(t, page, "<title>infra / @claude-prompt")
	assert.NotContains(t, page, "promptbox")
	assert.Contains(t, f.page(t, asViewer, "/runs"), "/runs/"+asked.ID)
	assert.NotContains(t, f.page(t, asOps, "/runs"), "/runs/"+asked.ID, "no prompt/view")
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/runs/"+asked.ID).Status)
}
