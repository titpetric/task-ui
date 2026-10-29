package view_test

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
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
		Secret:        &model.Secret{Scope: "all", Name: "TOKEN"},
		Account:       &model.User{ID: "u1", Name: "ops", Role: model.RoleRunner},
		Profile:       &view.Profile{Runs: []view.Run{run}, Prompts: []model.Note{{ID: "n1", Body: "why", Dir: "infra"}}, Matrix: &matrix},
		Matrices:      []view.Matrix{matrix},
		Workspace:     &model.Workspace{Folder: "malta"},
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
// reaches for a field that is not there fails here.
func TestView_Render(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	for _, name := range v.Names() {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, admin(t)), name)
		assert.Contains(t, buf.String(), "<!DOCTYPE html>", name)
		assert.Contains(t, buf.String(), "blacksky.digital", name)
		assert.Equal(t, 1, strings.Count(buf.String(), `<div class="toasts" id="toasts" role="status" aria-live="polite"></div>`), name)
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

// The Files list holds no Action or Summary column: what runs is in the
// Jobs table above it.
func TestView_RenderTree(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Tree = []folder.Item{{Name: "deploy.sh", Path: "deploy.sh", Href: "/files/infra/deploy.sh", Exec: true}, {Name: "docs", Dir: true}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "browse", p))
	out := buf.String()
	assert.Contains(t, out, `<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>`)
	assert.NotContains(t, out, "<th>Action</th>")
	assert.NotContains(t, out, "<th>Summary</th>")
	assert.NotContains(t, out, `data-label="Action"`)
	assert.Contains(t, out, `<span class="badge exec"`)
}

// The Jobs table has a row per job of its groups, under a heading row
// per group: the job's ID linked to its page, Run, with the field for
// args, and the confirm of a skill job, and a [...] of Run and
// summarize, the last run, History, and Run force for a job with hours.
// A job the user may not run has History; a running one Attach and Kill.
func TestView_RenderJobTable(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "folder", p))
		return buf.String()
	}
	table := func(p *view.Page) {
		f := p.Folder
		f.Hours = "Mon-Fri 9-17"
		build := &folder.Job{Folder: "infra", ID: "build", Command: []string{"make", "build"}}
		build.Name, build.Description = "build", "Build the images"
		custom := &folder.Job{Folder: "infra", ID: "@compose:custom", Command: []string{"atkins", "compose:custom"}}
		custom.Name, custom.TakesArgs = "custom", true
		up := &folder.Job{Folder: "infra", ID: "@compose:up", Command: []string{"atkins", "compose:up"}}
		up.Name = "up"
		script := &folder.Job{Folder: "infra", ID: "./deploy.sh", Kind: "file", Source: "deploy.sh", Command: []string{"./deploy.sh"}}
		script.Name = "deploy.sh"
		compose := &folder.Skill{ID: "compose", Project: "infra", Jobs: []*folder.Job{custom, up}}
		p.Actions = &view.Actions{Place: f, Self: "infra", Shell: f.ShellJob(), Skills: []*folder.Skill{compose}}
		p.JobGroups = []view.JobGroup{
			{Kind: "make", Note: "Makefile · make", Source: "Makefile", Folder: f, Jobs: []*folder.Job{build}, Targets: true},
			{Kind: "files", Folder: f, Jobs: []*folder.Job{script}},
			{Kind: "@compose", Folder: f, Skill: compose, Jobs: compose.Jobs},
			{Kind: "built-in", Folder: f, Jobs: []*folder.Job{f.ShellJob()}},
			{Kind: "make", Source: "Makefile", Dir: "tools", Folder: &folder.Folder{ID: "infra", Rel: "tools"}, Empty: "No targets in Makefile.", Targets: true},
		}
	}

	p := admin(t)
	p.Summaries = true
	table(p)
	p.Last = map[string]view.Run{
		"infra/build":       {Run: model.Run{ID: "r1", Folder: "infra", Job: "build", State: model.StateOK}},
		"infra/./deploy.sh": {Run: model.Run{ID: "r3", Folder: "infra", Job: "./deploy.sh", State: model.StateFailed}},
	}
	p.JobSummaries = map[string]*view.SummaryLink{"infra/build": view.NewSummaryLink("r1", "The build passed.")}
	out := render(p)
	// The dashboard's control component as named rows: a source's
	// targets under its file, linked to where they run, then the files,
	// the skill and the built-ins, each with its [...]; no columns, no
	// other.
	assert.Contains(t, out, `<table class="table control col-hover">`)
	assert.NotContains(t, out, `<th class="col">`)
	assert.Contains(t, out, `<a class="badge" href="/f/internal/infra">Makefile</a>`)
	assert.Contains(t, out, `<a class="badge" href="/f/internal/infra/tools">tools/Makefile</a>`)
	assert.Contains(t, out, `<span class="faint">No targets in Makefile.</span>`)
	assert.Equal(t, 5, strings.Count(out, `<tr class="named" data-span>`))
	for _, kind := range []string{"files", "@compose", "built-in"} {
		assert.Contains(t, out, `<span class="badge">`+kind+`</span>`, kind)
	}
	for _, id := range []string{"build", "./deploy.sh", "@compose:custom", "@compose:up", "@shell"} {
		assert.Contains(t, out, `data-job="`+id+`"`, id)
	}
	assert.Contains(t, out, `<button type="button" class="cc named ok" data-run data-folder="infra" data-job="build" data-title="infra / build" title="Run infra / build: Build the images">`)
	assert.Contains(t, out, `class="cc named bad" data-run data-folder="infra" data-job="./deploy.sh"`)
	assert.Contains(t, out, `class="cc named never" data-run data-folder="infra" data-job="@compose:custom"`)
	assert.Contains(t, out, `data-path="deploy.sh"`)
	assert.Contains(t, out, `data-args="atkins compose:custom --"`)
	assert.Contains(t, out, `data-confirm="Run compose up for infra?`)
	assert.Contains(t, out, `<span>Run and summarize build</span>`)
	assert.Contains(t, out, `<span>Run and summarize @compose:custom...</span>`)
	assert.Contains(t, out, `<span>Run force build</span>`)
	assert.Contains(t, out, `data-force="" title="Run out of hours, with a reason" data-run="" data-folder="infra" data-dir="" data-job="build"`)
	assert.Equal(t, 5, strings.Count(out, "data-force="), "Run force on every job, the built-ins too, in a workspace with hours")
	assert.Contains(t, out, `<span>View the last run of build</span>`)
	assert.Contains(t, out, `<span>History of @shell</span>`)
	assert.Contains(t, out, `href="/runs/r1/report" title="The build passed."`)
	assert.Equal(t, "/runs/r1/report", p.JobSummary("infra/build").Href)
	assert.Nil(t, p.JobSummary("infra/other"))

	viewer := admin(t)
	table(viewer)
	signIn(viewer, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	out = render(viewer)
	assert.NotContains(t, out, "data-run ")
	assert.Contains(t, out, `<a class="cc named idle never" href="/f/internal/infra/build" title="infra / build: Build the images: no permission to run it">`)

	running := admin(t)
	table(running)
	now := time.Now()
	run := view.Run{Run: model.Run{ID: "r2", Folder: "infra", Job: "build", State: model.StateRunning, StartedAt: &now}}
	running.Running = map[string]view.Run{"infra/build": run}
	out = render(running)
	assert.Contains(t, out, `<a class="cc named running" href="/runs/r2" data-attach="r2" title="Attach to infra / build, running">`)
	assert.Contains(t, out, `data-kill="r2"`)
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

// A job's page runs it with [ Run | v ], the caret opening the other
// ways to run it: Run and summarize and Run force. A user who may pin
// gets the job's Pin or Unpin at the head.
func TestView_RenderJobRunMenu(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Running = nil
	p.Summaries = true
	p.Pinning = true
	p.Folder.Hours = "mon-fri 9-17"
	render := func() string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "job", p))
		return buf.String()
	}

	out := render()
	assert.Contains(t, out, `<div class="run-split">`)
	assert.Contains(t, out, `<summary class="btn btn-primary btn-sm run-caret"`)
	assert.Contains(t, out, `<span>Run and summarize</span>`)
	assert.Contains(t, out, `<span>Run force</span>`)
	assert.Contains(t, out, `data-pin-folder="infra" data-pin-job="plan" data-pin-dir="" data-pin-key="infra||plan" title="Show it in the sidebar"`)
	assert.Contains(t, out, `>Pin plan</button>`)

	p.Pins = []model.Pin{{ID: "p1", Folder: "infra", Job: "plan"}}
	out = render()
	assert.Contains(t, out, `data-pin-id="p1"`)
	assert.Contains(t, out, `>Unpin plan</button>`)
}

// A run with a summary has a row under it holding the summary as
// rendered markdown.
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
	assert.Contains(t, out, `href="/f/internal/infra/plan">Job</a>`)
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

// A workspace's configure page opens Export in the dialog of the
// passphrase, typed or generated; the default workspace has no Delete.
func TestWorkspaceEdit_Page(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workspace = &model.Workspace{Folder: "malta"}
	p.KeysDir = "/keys"
	p.Form = map[string]string{"title": "Malta"}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	body := buf.String()
	assert.Contains(t, body, `<dialog class="confirm workspace-export" id="export-dialog"`)
	assert.Contains(t, body, `data-export="/api/workspaces/malta/export"`)
	assert.Contains(t, body, `data-passphrase-generate`)
	assert.Contains(t, body, `minlength="12"`)
	assert.Contains(t, body, `malta.tui.age`)

	p.Workspace = &model.Workspace{Folder: model.Default}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	assert.Contains(t, buf.String(), `data-dialog-open="export-dialog"`)
	assert.Contains(t, buf.String(), `data-delete="/api/workspaces/internal"`, "the default workspace deletes like any other")
	assert.NotContains(t, buf.String(), `Empty</span>`)
	assert.Contains(t, buf.String(), `data-confirm-ok="Delete"`)
	assert.Contains(t, buf.String(), `Delete the default workspace internal with its keys, its home? The root&#39;s built-in jobs keep running; Add project then needs a workspace.`)
}

// A workspace's settings page shows its home: an empty one with Clone,
// a checkout with its repository, Replace and Remove.
func TestWorkspaceEdit_Home(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workspace = &model.Workspace{Folder: "malta", Title: "Malta"}
	p.Workspaces = map[string]int{"malta": 1}
	p.Form = map[string]string{"title": "Malta"}
	p.Home = &view.WorkspaceHome{Folder: "malta/home", Dir: "/data/malta/home"}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	body := buf.String()
	assert.Contains(t, body, `<h3>Home repository <small>malta/home/, the HOME of the workspace&#39;s runs</small></h3>`)
	assert.Contains(t, body, `<p class="empty-line">The home is an empty directory.</p>`)
	assert.Contains(t, body, `data-dialog-open="home-replace-dialog">`)
	assert.Contains(t, body, `>Clone</button>`)
	assert.NotContains(t, body, `>Remove</button>`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/home" data-method="PUT"`)

	p.Home.URL, p.Home.Branch = "git@example.com:malta/workspace.git", "main"
	p.Home.Checkout = &git.Info{Remote: p.Home.URL, Branch: "main", Commit: "abc1234", Subject: "Skills"}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	body = buf.String()
	assert.NotContains(t, body, `The home is an empty directory.`)
	assert.Contains(t, body, `git@example.com:malta/workspace.git`)
	assert.Contains(t, body, `href="/f/malta/home"`)
	assert.Contains(t, body, `abc1234`)
	assert.Contains(t, body, `>Replace</button>`)
	assert.Contains(t, body, `>Remove</button>`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/home" data-method="DELETE"`)
}

// The Workspaces page opens Import in a dialog from its head.
func TestWorkspaces_ImportDialog(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workspaces", admin(t)))
	body := buf.String()
	assert.Contains(t, body, `data-dialog-open="import-dialog"`)
	assert.Contains(t, body, `<dialog class="confirm workspace-import" id="import-dialog"`)
	assert.Contains(t, body, `data-import="/api/workspaces/import"`)
}

// A folder page has no Environment section: variables live on the
// Secrets page alone.
func TestView_RenderEnvironment(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "folder", admin(t)))
	assert.NotContains(t, buf.String(), "Environment")
	assert.NotContains(t, buf.String(), "env file")
	assert.NotContains(t, buf.String(), "/env/")
}
