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
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/permission"
)

// admin is a page as an admin sees it, with one folder and one run.
//
// It is the page model the whole package renders over, so a test of what
// one component draws sits in that component's file: the Jobs table in
// job_group_test.go, the sidebar in sidebar_test.go, the head's menus in
// head_test.go, the roles in role_members_test.go and a user's keys in
// user_keys_test.go.
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

	// The one test that builds its own: this is the constructor under
	// test, and the rest of the package shares what it made.
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
// reaches for a field that is not there fails here. What it draws is not
// asserted: a template is read, and the markup tests that read it are
// behind TASK_UI_VIEW_TESTS. This one builds its own set so it runs
// whatever that says.
func TestView_Render(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	for _, name := range v.Names() {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, admin(t)), name)
		assert.True(t, buf.Len() > 0, name)
	}

	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "login", &view.Page{Back: "/"}))
	assert.True(t, buf.Len() > 0)
	assert.Error(t, v.Render(&buf, "nothing", admin(t)))
}

// The window carries the reason field a run out of hours requires.
func TestView_RenderDialog(t *testing.T) {
	t.Parallel()

	v := shared(t)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", admin(t)))
	assert.Contains(t, buf.String(), `id="run-reason-text"`)
	assert.Contains(t, buf.String(), `id="run-force"`)
}

// The Files list holds no Action or Summary column: what runs is in the
// Jobs table above it.
func TestView_RenderTree(t *testing.T) {
	t.Parallel()

	v := shared(t)
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

// A job's runs have Summarize for a user who may, on a run that ended
// without a summary; a summary under way says so in its place.
func TestView_RenderJobSummarize(t *testing.T) {
	t.Parallel()

	v := shared(t)
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

	v := shared(t)
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

	v := shared(t)
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

	v := shared(t)
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

	v := shared(t)
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

	v := shared(t)
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
	v := shared(t)
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

// The Workspaces page opens Import in a dialog from its head.
func TestWorkspaces_ImportDialog(t *testing.T) {
	t.Parallel()
	v := shared(t)
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

	v := shared(t)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "folder", admin(t)))
	assert.NotContains(t, buf.String(), "Environment")
	assert.NotContains(t, buf.String(), "env file")
	assert.NotContains(t, buf.String(), "/env/")
}

// The keyboard's first stop is a link past the sidebar to the content,
// which takes the focus; the header's prompt and its logo button sit on
// the Tab path.
func TestLayout_Keyboard(t *testing.T) {
	t.Parallel()

	body := render(t, "home", admin(t))
	skip := strings.Index(body, `<a class="skip-link" href="#main">Skip to content</a>`)
	header := strings.Index(body, `<header class="top"`)
	main := strings.Index(body, `<main id="main" tabindex="-1">`)
	assert.True(t, skip > 0 && skip < header && header < main, "skip link, header, then main")
	head := body[header:strings.Index(body, `</header>`)]
	assert.NotContains(t, head, `tabindex="-1"`)
	assert.Contains(t, head, `id="header-toggle" aria-label="Prompt"`)
	assert.Contains(t, head, `id="header-prompt" aria-label="Ask Claude"`)

	login := &view.Page{Title: "Sign in"}
	assert.NotContains(t, render(t, "login", login), "skip-link", "no sidebar on the sign in screen")
}

// Add workspace lands on the new workspace, where its project and its
// keys are one click away.
func TestWorkspaceNew_LandsOnTheWorkspace(t *testing.T) {
	t.Parallel()

	body := render(t, "workspace_new", admin(t))
	assert.Contains(t, body, `data-api="/api/workspaces" data-method="POST" data-next="/settings/workspaces/{folder}?notice=Workspace+added"`)
}

// workspacesRow returns the row of the workspace in the rendered list.
func workspacesRow(t *testing.T, body, folder string) string {
	t.Helper()
	start := strings.Index(body, `<a href="/f/`+folder+`">`)
	assert.True(t, start >= 0, "a row for "+folder)
	end := strings.Index(body[start:], "</tr>")
	return body[start : start+end]
}

// The row of a workspace adds a project, lists its keys and its secrets
// without opening Edit workspace first.
func TestWorkspaces_RowMenu(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.WorkspaceList = []model.Workspace{{Folder: "malta", Title: "Malta"}}
	p.Workspaces = map[string]int{"malta": 0}
	row := workspacesRow(t, render(t, "workspaces", p), "malta")
	assert.Contains(t, row, `href="/settings/projects/new?workspace=malta"`)
	assert.Contains(t, row, `href="/settings/workspaces/malta/keys"`)
	assert.Contains(t, row, `href="/settings/secrets?workspace=malta"`)
	add := strings.Index(row, "Add project")
	keys := strings.Index(row, "SSH keys")
	secrets := strings.Index(row, "<span>Secrets</span>")
	edit := strings.Index(row, "Edit workspace")
	assert.True(t, edit < add && add < keys && keys < secrets, "Edit workspace, Add project, SSH keys, Secrets")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	row = workspacesRow(t, render(t, "workspaces", p), "malta")
	assert.NotContains(t, row, "Add project", "adding a project takes project/edit")
	assert.NotContains(t, row, "<span>Secrets</span>", "the secrets take secret/view")
	assert.Contains(t, row, "SSH keys")
}

// A workspace with no project, or no secret, says so and links the form
// that adds one in the cell, for a user who may; with some, the count links.
func TestWorkspaces_EmptyCounts(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.WorkspaceList = []model.Workspace{{Folder: "empty"}, {Folder: "full"}}
	p.Workspaces = map[string]int{"empty": 0, "full": 1}
	p.Nested = map[string][]string{"full": {"full/api", "full/web"}}
	p.SecretCounts = map[string]int{"full": 2}
	body := render(t, "workspaces", p)

	empty := workspacesRow(t, body, "empty")
	assert.Contains(t, empty, `<a href="/settings/projects/new?workspace=empty" aria-label="Add a project to empty">add</a>`)
	assert.Contains(t, empty, `<a href="/settings/secrets?workspace=empty" aria-label="Add a secret to empty">add</a>`)

	full := workspacesRow(t, body, "full")
	assert.Contains(t, full, `<a href="/settings/workspaces/full" title="full/api, full/web">2</a>`)
	assert.NotContains(t, full, "aria-label=\"Add a project")
	assert.NotContains(t, full, "aria-label=\"Add a secret")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow})
	empty = workspacesRow(t, render(t, "workspaces", p), "empty")
	assert.NotContains(t, empty, "aria-label=\"Add a")
	assert.Contains(t, empty, `<span class="faint">none</span>`)
}
