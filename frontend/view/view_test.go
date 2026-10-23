package view_test

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/permission"
)

// admin is a page as an admin sees it, with one folder and one run.
func admin(t *testing.T) *view.Page {
	t.Helper()
	now := time.Now()
	f := &folder.Folder{ID: "infra", Title: "Infra"}
	job := &folder.Job{Folder: "infra", ID: "plan", Command: []string{"make", "plan"}}
	job.Name = "plan"
	f.Jobs = []*folder.Job{job}
	run := view.Run{Run: model.Run{ID: model.NewID(), Folder: "infra", Job: "plan", UserName: "ops", State: model.StateRunning, StartedAt: &now}, Live: true}
	user := &model.User{Name: "admin", Role: model.RoleAdmin}
	matrix := view.NewMatrix(config.Default(), permission.Policy{"run": permission.Deny}, nil, false)
	matrix.Editable, matrix.Effective = true, true
	return &view.Page{
		Title:         "Test",
		Path:          "/",
		User:          user,
		Permissions:   permission.Set{User: user},
		LoginRequired: true,
		Folders:       []*folder.Folder{f},
		Recents:       []view.RecentGroup{{Label: "In the last hour", Entries: []view.Recent{{Run: &run}}}},
		Running:       map[string]view.Run{run.Key(): run},
		Folder:        f,
		Job:           job,
		Run:           &run,
		Runs:          []view.Run{run},
		EnvFile:       &envfile.File{Folder: "infra", Path: ".env", Content: "A=1\n"},
		Secret:        &model.Secret{Scope: "all", Name: "TOKEN"},
		Account:       &model.User{ID: "u1", Name: "ops", Role: model.RoleRunner},
		Profile:       &view.Profile{Runs: []view.Run{run}, Prompts: []model.Note{{ID: "n1", Body: "why", Dir: "infra"}}, Matrix: &matrix},
		Matrices:      []view.Matrix{matrix},
		Workspace:     &model.Workspace{Folder: "malta", SourceURL: "git@example.com:org/workspace.git"},
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	names := v.Names()
	assert.Contains(t, names, "home")
	assert.Contains(t, names, "login")
	assert.Contains(t, names, "folder")
	assert.Contains(t, names, "job")
	assert.Contains(t, names, "runs")
	assert.Contains(t, names, "run")
	assert.Contains(t, names, "artifacts")
	assert.Contains(t, names, "schedules")
	assert.Contains(t, names, "settings")
	assert.Contains(t, names, "notfound")
	assert.Contains(t, names, "forbidden")
	assert.NotContains(t, names, "_layout")
}

// Every page renders over the same page model, so a template that
// reaches for a field that is not there fails here, not in a browser.
func TestView_Render(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	for _, name := range v.Names() {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, admin(t)), name)
		assert.Contains(t, buf.String(), "<!DOCTYPE html>", name)
		assert.Contains(t, buf.String(), "blacksky.digital", name)
		if name != "login" {
			assert.Contains(t, buf.String(), `class="avatar role-admin"`, name)
			assert.Contains(t, buf.String(), "Recents", name)
		}
	}

	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "login", &view.Page{Back: "/"}))
	assert.Contains(t, buf.String(), "Sign in")
	assert.NotContains(t, buf.String(), "Recents")

	assert.Error(t, v.Render(&buf, "nothing", admin(t)))
}

// The window carries the reason field a run out of hours requires.
func TestView_RenderDialog(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", admin(t)))
	assert.Contains(t, buf.String(), `id="run-reason-text"`)
	assert.Contains(t, buf.String(), `id="run-force"`)
}

// The Files list has an Action column, Run, and a Summary column,
// |>summarize for a user who may and the link to the summary of the
// row's last run.
func TestView_RenderTree(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "browse", p))
		return buf.String()
	}
	tree := func(summaries bool) *view.Page {
		p := admin(t)
		build := &folder.Job{Folder: "infra", ID: "build", Command: []string{"make", "build"}}
		build.Name = "build"
		p.Folder.Jobs = append(p.Folder.Jobs, build)
		p.Summaries = summaries
		p.Tree = []folder.Item{{Name: "Makefile", Path: "Makefile", Href: "/f/infra/Makefile"}}
		p.TreeMenus = map[string]*view.RunMenu{"Makefile": view.NewRunMenu(p.Folder, []*folder.Job{build}, build)}
		return p
	}

	plain := render(tree(false))
	assert.Contains(t, plain, "<th>Action</th><th>Summary</th>")
	assert.Contains(t, plain, `data-label="Action"`)
	assert.Contains(t, plain, `data-label="Summary"`)
	assert.Contains(t, plain, `data-job="build"`)
	assert.NotContains(t, plain, "data-summarize", "no prompt/edit")
	assert.NotContains(t, plain, "summary-link")

	p := tree(true)
	p.TreeSummaries = map[string]*view.SummaryLink{"Makefile": view.NewSummaryLink("r1", "The build passed.")}
	out := render(p)
	assert.Contains(t, out, "data-summarize")
	assert.Contains(t, out, `href="/runs/r1/report" title="The build passed."`)
	assert.Equal(t, "/runs/r1/report", p.TreeSummary("Makefile").Href)
	assert.Nil(t, p.TreeSummary("other"))
}

// A job's runs have Summarize for a user who may, on a run that ended
// without a summary; a summary under way says so in its place.
func TestView_RenderJobSummarize(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "job", p))
		return buf.String()
	}
	p := admin(t)
	p.Runs[0].State = model.StateOK
	assert.NotContains(t, render(p), "data-summarize-run", "no Summaries")

	p.Summaries = true
	out := render(p)
	assert.Contains(t, out, `data-summarize-run="`+p.Runs[0].ID+`"`)
	assert.Contains(t, out, `data-question="Summarize this run."`)

	p.RunReports = map[string]*note.Report{p.Runs[0].ID: {RunID: p.Runs[0].ID, Pending: true}}
	out = render(p)
	assert.NotContains(t, out, "data-summarize-run")
	assert.Contains(t, out, ">summarizing<")

	p.RunReports = nil
	p.Runs[0].State = model.StateRunning
	assert.NotContains(t, render(p), "data-summarize-run", "still running")
}

// A run with a summary has a row under it holding the summary as
// rendered markdown, not as plain text.
func TestView_RenderJobSummary(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.RunReports = map[string]*note.Report{p.Runs[0].ID: {RunID: p.Runs[0].ID, By: "ops", Body: "## Passed\n\n- 3 images"}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "job", p))
	out := buf.String()
	assert.Contains(t, out, `class="summary-row"`)
	assert.Contains(t, out, "<li>3 images</li>")
	assert.Contains(t, out, "asked by ops")
	assert.NotContains(t, out, `<pre class="summary">## Passed`)
}

// The report page shows Claude's markdown as a document, with links to
// the run, Claude's run and the job, and no terminal; a question that
// is no summary's shows above its answer.
func TestView_RenderReport(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "report", p))
		return buf.String()
	}
	p := admin(t)
	about := p.Runs[0]
	about.State = model.StateOK
	claude := view.Run{Run: model.Run{ID: "c1", Folder: "infra", Job: model.PromptJob, UserName: "ops", State: model.StateOK}}
	p.Run, p.ReportOf = &claude, &about
	p.Report = &note.Report{RunID: "c1", AboutRunID: about.ID, ThreadID: "t", Question: note.SummaryQuestion, By: "ops", Body: "## Passed\n\n<script>x</script>"}
	out := render(p)
	assert.Contains(t, out, "Claude's summary of the run")
	assert.Contains(t, out, "<h3>Summary</h3>")
	assert.Contains(t, out, `href="/runs/`+about.ID+`">Run</a>`)
	assert.Contains(t, out, `href="/runs/c1">Claude's run</a>`)
	assert.Contains(t, out, `href="/f/_/infra/plan">Job</a>`)
	assert.Contains(t, out, "&lt;script&gt;")
	assert.NotContains(t, out, `data-output=`)
	assert.NotContains(t, out, "<h3>Question</h3>")

	p.Report.Question = "why so slow?"
	out = render(p)
	assert.Contains(t, out, "<h3>Question</h3>")
	assert.Contains(t, out, "<h3>Answer</h3>")
}

// A run with a report links it from its page and from its row's
// summary on the job page.
func TestView_RenderReportLinks(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	id := p.Runs[0].ID
	p.Report = &note.Report{RunID: id, Body: "done"}
	p.RunReports = map[string]*note.Report{id: p.Report}
	for _, name := range []string{"run", "job"} {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, p), name)
		assert.Contains(t, buf.String(), `href="/runs/`+id+`/report">Report</a>`, name)
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", p))
	assert.Contains(t, buf.String(), `id="run-report"`)
}

// Every data table on every page has a heading of its own since the
// table before it: an h3 for a block, an h4 inside a block of more, or
// the label of the form field a table of inputs is.
func TestView_RenderTableHeadings(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	heading := regexp.MustCompile(`<h[34][ >]|<span class="label">`)
	for _, name := range v.Names() {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, admin(t)), name)
		parts := strings.Split(buf.String(), `<table class="table`)
		for i := 1; i < len(parts); i++ {
			before := parts[i-1]
			if at := strings.LastIndex(before, "</table>"); at >= 0 {
				before = before[at:]
			}
			assert.True(t, heading.MatchString(before), name+": table "+strconv.Itoa(i)+" has no heading")
		}
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "job", admin(t)))
	assert.Contains(t, buf.String(), "<h4>Runs <small>newest first</small></h4>")
}
