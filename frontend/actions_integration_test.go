//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills/skillstest"
)

// actionsBar cuts the actions bar out of a page.
func actionsBar(body string) string {
	_, rest, ok := strings.Cut(body, `<div class="actionsbar"`)
	if !ok {
		return ""
	}
	bar, _, _ := strings.Cut(rest, "\n</div>")
	return bar
}

// jobsTable cuts the Jobs table out of a page.
func jobsTable(body string) string {
	_, rest, ok := strings.Cut(body, `<table class="table control col-hover">`)
	if !ok {
		return ""
	}
	table, _, _ := strings.Cut(rest, "</table>")
	return table
}

// A folder page and a directory browsed inside a folder carry the git
// line of the checkout, and in the Jobs table a group per skill atkins
// applies there, each job a Run of the folder in the directory the job
// runs in, beside the place's own atkins jobs. A viewer gets History
// instead; compose up and down of task-ui's own project ask first.
func TestFolder_Actions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	probe, _ := skillstest.Prober(t)
	f.infra.Probe = probe
	out, err := exec.Command("git", "-c", "init.defaultBranch=main", "init", "-q", f.infra.Dir).CombinedOutput()
	assert.NoError(t, err, string(out))
	app := filepath.Join(f.infra.Dir, "app")
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "compose.yml"), []byte("services:\n  a:\n    build: .\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "atkins.yml"), []byte("jobs:\n  default:\n    steps:\n      - run: echo hi\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "new.txt"), []byte("x\n"), 0o644))

	topPage := f.page(t, asOps, "/f/_/infra")
	top := actionsBar(topPage)
	assert.Contains(t, top, `<span class="glyph" aria-hidden="true">⎇</span> main`)
	assert.Contains(t, top, `●2`, "new.txt and app/")
	assert.Contains(t, top, `· src/infra`, "the checkout under the root")
	assert.NotContains(t, top, "data-run", "the bar is the git line alone")
	topJobs := jobsTable(topPage)
	assert.Contains(t, topJobs, `<span class="badge">@git</span>`)
	assert.Contains(t, topJobs, `data-folder="infra" data-job="@git:pull" data-title="infra / pull"`)
	assert.NotContains(t, topJobs, "@compose")

	appPage := f.page(t, asOps, "/f/_/infra/app")
	assert.Contains(t, actionsBar(appPage), `⎇</span> main`)
	bar := jobsTable(appPage)
	assert.Contains(t, bar, `data-folder="infra" data-job="@git:status" data-title="infra / status"`, "git runs in the checkout")
	assert.Contains(t, bar, `data-folder="infra" data-dir="app" data-job="@compose:build" data-title="infra/app / build"`)
	assert.Contains(t, bar, `data-dir="app" data-job="@compose:up"`)
	// compose custom takes args: its Run opens a field naming what they
	// follow first.
	assert.Contains(t, bar, `data-job="@compose:custom" data-title="infra/app / custom" data-args="docker compose" title=`)
	assert.Contains(t, bar, `, with arguments after docker compose"`)
	assert.NotContains(t, bar, `data-job="@compose:up" data-title="infra/app / up" data-args`)
	assert.Contains(t, bar, `title="atkins `)
	assert.Contains(t, f.page(t, asOps, "/f/_/infra/app"), "Its jobs run in infra/app/")
	deploy := filepath.Join(f.infra.Dir, "deploy")
	assert.NoError(t, os.MkdirAll(deploy, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(deploy, "compose.yml"), []byte("services:\n  a:\n    image: x\n"), 0o644))
	skillsOnly := f.page(t, asOps, "/f/_/infra/deploy")
	assert.Contains(t, skillsOnly, "the jobs below run here")
	assert.NotContains(t, jobsTable(skillsOnly), "@compose:build")
	assert.Contains(t, bar, `data-folder="infra" data-dir="app" data-job="default"`)
	assert.NotContains(t, bar, "data-confirm")

	viewer := jobsTable(f.page(t, asViewer, "/f/_/infra/app"))
	assert.Contains(t, viewer, `<a class="cc named idle never" href="/f/_/infra/app/@compose:up"`)
	assert.Contains(t, viewer, `<a class="cc named idle never" href="/f/_/infra/@git:status"`)
	assert.Contains(t, viewer, `: no permission to run it">`)
	assert.NotContains(t, viewer, "data-run")
	page := f.page(t, asViewer, "/f/_/infra/app/@compose:up")
	assert.Contains(t, page, "infra/app / up")
	assert.Contains(t, page, "atkins compose:up")
	assert.Contains(t, f.page(t, asViewer, "/f/_/infra/@git:status"), "Show the branch")

	// The project task-ui runs in: up and down ask first.
	f.opts.ComposeProject = "app"
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
	bar = jobsTable(f.page(t, asOps, "/f/_/infra/app"))
	assert.Contains(t, bar, `data-job="@compose:up" data-title="infra/app / up" data-confirm="Run compose up for app?`)
	assert.Contains(t, bar, `data-job="@compose:down" data-title="infra/app / down" data-confirm="Run compose down for app?`)
	assert.NotContains(t, bar, `data-job="@compose:build" data-title="infra/app / build" data-confirm`)

	// No checkout: no git line.
	assert.Equal(t, "", actionsBar(f.page(t, asOps, "/f/_/private")))
}

// An admin with a Claude token gets the prompt box on every folder page
// and on a directory browsed inside a folder, posting its place to
// /api/prompts; open on a workspace's page, folded elsewhere.
func TestFolder_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	const block = `data-api="/api/prompts" data-refresh-notes`
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))

	// infra is a workspace; without a Claude token there is no prompt.
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "infra", Title: "Infra"}))
	assert.NotContains(t, f.page(t, asAdmin, "/f/infra"), block)

	// The token reaches everyone; the prompt is the admins'.
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	admin := f.page(t, asAdmin, "/f/infra")
	assert.Contains(t, admin, `<details class="promptbox" open>`)
	assert.Contains(t, admin, block)
	assert.Contains(t, admin, `<input type="hidden" name="folder" value="infra">`)
	assert.Contains(t, admin, `<textarea name="prompt"`)
	assert.NotContains(t, f.page(t, asOps, "/f/infra"), block)
	assert.NotContains(t, f.page(t, asViewer, "/f/infra"), block)

	// A folder that is not a workspace takes one too, folded, and so
	// does a directory inside a folder.
	private := f.page(t, asAdmin, "/f/_/private")
	assert.Contains(t, private, `<details class="promptbox">`)
	assert.Contains(t, private, `<input type="hidden" name="folder" value="private">`)
	browse := f.page(t, asAdmin, "/f/infra/images")
	assert.Contains(t, browse, block)
	assert.Contains(t, browse, `<input type="hidden" name="dir" value="images">`)
	assert.Contains(t, browse, "Claude Code in infra/images/")
	assert.NotContains(t, f.page(t, asOps, "/f/infra/images"), block)

	// An admin picks who the prompt runs as, the login selected, in the
	// box and in the header; a runner with prompt/edit prompts as themself.
	assert.Contains(t, admin, `Runs @claude-prompt as <select name="user" aria-label="Runs as">`)
	assert.Contains(t, admin, `<option value="admin" selected data-keys="">admin · no keys</option>`)
	assert.Contains(t, admin, `<option value="ops" data-keys="">ops · no keys</option>`)
	assert.Contains(t, admin, `id="header-prompt-user"`)
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asOps, "prompt/edit", "allow"))
	ops := f.page(t, asOps, "/f/infra")
	assert.Contains(t, ops, block)
	assert.Contains(t, ops, `Runs @claude-prompt as you in infra/`)
	assert.NotContains(t, ops, `name="user"`)
	assert.NotContains(t, ops, `header-prompt-user`)
}

// @shell is a built-in row of the Jobs table of a folder and of a
// directory inside one for an admin, and for no one else.
func TestFolder_Shell(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))
	const button = `data-job="@shell"`

	admin := f.page(t, asAdmin, "/f/_/infra")
	assert.Contains(t, admin, button)
	assert.Contains(t, admin, `data-title="infra / shell"`)
	browse := f.page(t, asAdmin, "/f/_/infra/images")
	assert.Contains(t, browse, `data-dir="images" data-job="@shell" data-title="infra/images / shell"`)
	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.NotContains(t, f.page(t, user, "/f/_/infra"), button, user)
		assert.NotContains(t, f.page(t, user, "/f/_/infra/images"), button, user)
	}
}

// @claude follows @shell among the built-ins, for an admin with a
// Claude token: @claude of the folder or of the directory inside it.
func TestFolder_Claude(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))
	const button = `data-job="@claude"`

	assert.NotContains(t, f.page(t, asAdmin, "/f/_/infra"), button, "no Claude token")
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	admin := f.page(t, asAdmin, "/f/_/infra")
	assert.Contains(t, admin, `data-folder="infra" data-job="@claude" data-title="infra / claude"`)
	assert.Contains(t, admin, `href="/f/_/infra/@claude"><svg class="ic"><use href="#i-list"/></svg><span>History of @claude</span></a>`)
	assert.Contains(t, admin, `href="/f/_/infra/@shell"><svg class="ic"><use href="#i-list"/></svg><span>History of @shell</span></a>`)
	browse := f.page(t, asAdmin, "/f/_/infra/images")
	assert.Contains(t, browse, `data-dir="images" data-job="@claude" data-title="infra/images / claude"`)
	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.NotContains(t, f.page(t, user, "/f/_/infra"), button, user)
	}
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/f/_/infra/@claude").Status, "its job page")
}

// An admin's [...] menu in the page head links the settings of the
// workspace or the project a folder page is of; no one else's does, nor
// a directory's inside them.
func TestFolder_SettingsLinks(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "infra", Title: "Infra"}))
	assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: "private", Title: "Private"}))
	const workspace, project = `<span>Edit workspace</span>`, `<span>Edit project</span>`

	ws := f.page(t, asAdmin, "/f/infra")
	assert.Contains(t, ws, `href="/settings/workspaces/infra"`)
	assert.Contains(t, ws, workspace)
	assert.NotContains(t, ws, project)
	root := f.page(t, asAdmin, "/f/_")
	assert.Contains(t, root, `href="/settings/workspaces/_"`)
	pr := f.page(t, asAdmin, "/f/_/private")
	assert.Contains(t, pr, `href="/settings/projects/private"`)
	assert.Contains(t, pr, project)
	assert.NotContains(t, pr, workspace)
	browse := f.page(t, asAdmin, "/f/infra/images")
	assert.NotContains(t, browse, workspace)
	assert.NotContains(t, browse, project)
	for _, user := range []string{asOps, asViewer} {
		for _, path := range []string{"/f/infra", "/f/_", "/f/_/private"} {
			page := f.page(t, user, path)
			assert.NotContains(t, page, workspace, user+" "+path)
			assert.NotContains(t, page, project, user+" "+path)
		}
	}
}

// |> Summarize stands beside Run on the job's page, and in the row's
// menu of the folder's Jobs table, for a user with prompt/edit and a
// Claude token; no one else gets it, and neither does an interactive
// job.
func TestFolder_Summarize(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	const button = `data-run data-summarize data-folder="infra" data-job="plan" data-title="infra / plan" title="Run and summarize">|&gt; Summarize</button>`

	assert.NotContains(t, f.page(t, asAdmin, "/f/_/infra"), "data-summarize", "no Claude token")
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	assert.Contains(t, f.page(t, asAdmin, "/f/_/infra"), `data-summarize="" data-run="" data-folder="infra" data-dir="" data-job="plan" data-title="infra / plan"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize plan</span>`)
	assert.Contains(t, f.page(t, asAdmin, "/f/_/infra/plan"), button)
	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.NotContains(t, f.page(t, user, "/f/_/infra"), "data-summarize", user)
	}
	f.infra.Job("many").Interactive = true
	assert.NotContains(t, f.page(t, asAdmin, "/f/_/infra"), "Run and summarize many")
}
