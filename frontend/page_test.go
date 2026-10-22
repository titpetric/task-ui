package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The chrome every page shares: the user, the folders they see, and the
// recent runs with who started them.
func TestPage_Chrome(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	plan := f.run(t, "infra", "plan", "ops", model.StateOK)
	f.run(t, "private", "deploy", "admin", model.StateFailed)

	body := f.page(t, asAdmin, "/")
	// A recent is a link to its run's page; the run window opens from
	// View and Attach, and links there in its header.
	recents := body[strings.Index(body, `id="recents"`):]
	assert.Contains(t, recents, `<a href="/runs/`+plan.ID+`" class="recent" title="Open infra / plan, ops, ok">`)
	assert.NotContains(t, recents, "data-attach")
	assert.Contains(t, body, `<a class="btn btn-sm" id="run-page" href="/runs" hidden>Open run page</a>`)
	assert.Contains(t, body, `class="avatar role-admin">AD<`)
	assert.Contains(t, body, "Settings")
	assert.Contains(t, body, "Logout")
	assert.Contains(t, body, `<a href="/users/admin"><svg class="ic"><use href="#i-users"/></svg>Profile</a>`)
	assert.Contains(t, body, `data-folder="infra"`)
	assert.Contains(t, body, `data-folder="private"`)
	assert.Contains(t, body, "Recents")
	assert.Contains(t, body, `class="avatar sm role-runner" title="ops">OP<`)
	assert.Contains(t, body, `class="avatar sm role-admin" title="admin">AD<`)
	// The body carries each user's role by name, and nothing else, for
	// the avatars the script draws.
	assert.Contains(t, body, `<body data-roles='{&#34;admin&#34;:&#34;admin&#34;,&#34;limited&#34;:&#34;runner&#34;,&#34;ops&#34;:&#34;runner&#34;,&#34;viewer&#34;:&#34;viewer&#34;}' data-workspaces='{}' data-me="admin" data-user-menu="edit">`)

	// The header's prompt, for prompt/edit: a bare field with nothing
	// that names it.
	const prompt = `<input type="text" class="header-prompt" id="header-prompt" tabindex="-1" autocomplete="off" spellcheck="false">`
	assert.Contains(t, body, prompt)
	head := body[strings.Index(body, `<header class="top">`):strings.Index(body, `</header>`)]
	assert.NotContains(t, head, "placeholder")
	assert.NotContains(t, head[strings.Index(head, "header-prompt"):strings.Index(head, `<details class="user"`)], "aria-label")

	limited := f.page(t, asLimited, "/")
	assert.NotContains(t, limited, "header-prompt")
	assert.NotContains(t, f.page(t, asViewer, "/"), "header-prompt")
	assert.NotContains(t, limited, "Settings</a>")
	assert.NotContains(t, limited, `data-folder="private"`)
	assert.NotContains(t, limited, `title="admin">AD<`)
	assert.Contains(t, limited, `title="ops">OP<`)
}

func mustProject(t *testing.T, f *fixture, folder string) *model.Project {
	t.Helper()
	project, err := f.opts.Projects.GetByFolder(context.Background(), folder)
	assert.NoError(t, err)
	return project
}

// A project in the sidebar folds: a disclosure with the folders inside
// it, open by default, the chevron on its row.
func TestPage_ProjectsFold(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/")
	infra := strings.Index(body, `<details class="project" data-project="infra" open>`)
	edge := strings.Index(body, `<details class="folder" data-folder="infra/edge"`)
	private := strings.Index(body, `<details class="project" data-project="private" open>`)
	assert.True(t, infra >= 0 && edge > infra && private > edge, "infra holds infra/edge, then private")
	assert.True(t, strings.Index(body[infra:], "</details>") > edge-infra, "infra closes after infra/edge")
	chev := body[infra : strings.Index(body[infra:], "</summary>")+infra]
	assert.Contains(t, chev, `<a class="section" href="/f/_/infra" data-folder="infra"`)
	assert.Contains(t, chev, `<svg class="ic chev"><use href="#i-chev"/></svg>`)
	// Without a workspace the sidebar has none to fold.
	assert.NotContains(t, body, `class="workspace`)

	// The sidebar leaves out a project set not to show there, with the
	// folders inside it; its pages stay.
	assert.NoError(t, f.opts.Projects.Create(context.Background(), &model.Project{Folder: "infra", SidebarHidden: true}))
	hidden := f.page(t, asAdmin, "/")
	assert.NotContains(t, hidden, `data-project="infra"`)
	assert.NotContains(t, hidden, `<details class="folder" data-folder="infra/edge"`)
	assert.Contains(t, hidden, `data-project="private"`)
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/f/_/infra").Status)
	assert.NoError(t, f.opts.Projects.Delete(context.Background(), mustProject(t, f, "infra").ID))

	// The sidebar marks the open project, on the project and its whole row.
	page := f.page(t, asAdmin, "/f/_/infra")
	assert.Contains(t, page, `<details class="project active" data-project="infra" open>`)
	assert.Contains(t, page, `<summary class="active" title="infra/`)
	assert.Contains(t, page, `<a class="section active" href="/f/_/infra"`)
	// A folder inside it keeps the project open, and lights its own row only.
	edgePage := f.page(t, asAdmin, "/f/_/infra/edge")
	assert.Contains(t, edgePage, `<details class="project active" data-project="infra" open>`)
	assert.NotContains(t, edgePage, `<summary class="active"`)
}

func TestPage_NoticeAndError(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/runs?notice=Saved&error=Oops")
	assert.Contains(t, body, `<div class="alert ok" role="status">Saved</div>`)
	assert.Contains(t, body, `<div class="alert bad" role="alert">Oops</div>`)
}

func TestPage_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, path := range []string{"/settings", "/settings/projects", "/settings/keys", "/settings/secrets", "/settings/users", "/f/_/infra/env/.env"} {
		res := f.get(t, asOps, path)
		assert.Equal(t, http.StatusForbidden, res.Status, path)
		assert.Contains(t, res.Body, "Settings are for admins.")
	}
}

// The internal folder is the admins': in the sidebar, on the dashboard
// and at /f/internal, with its three jobs. Nobody else finds it.
func TestPage_internal(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.opts.Catalogue.(catalogue)[model.Internal] = agent.Folder(f.root, "claude")

	home := f.page(t, asAdmin, "/")
	assert.Contains(t, home, `href="/f/internal"`)
	assert.Contains(t, home, "claude-note")
	page := f.page(t, asAdmin, "/f/internal")
	for _, job := range []string{"claude-note", "claude-prompt", "claude"} {
		assert.Contains(t, page, `href="/f/internal/`+job+`"`)
		assert.Contains(t, f.page(t, asAdmin, "/f/internal/"+job), "Claude Code")
	}

	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.NotContains(t, f.page(t, user, "/"), `/f/internal`, user)
		assert.Equal(t, http.StatusNotFound, f.get(t, user, "/f/internal").Status, user)
		assert.Equal(t, http.StatusNotFound, f.get(t, user, "/f/internal/claude-note").Status, user)
	}
}

// A workspace in the sidebar folds its projects under it, each a project
// with its own fold and its folders; the fold keys are the full IDs and
// the links carry the namespace. Until an admin makes the workspace, its
// checkout is a project like any other.
func TestPage_Workspaces(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	cat := f.opts.Catalogue.(catalogue)
	for _, id := range []string{"malta", "malta/deploy", "malta/deploy/web", "malta/apps"} {
		cat[id] = shellFolder(t, f.root, id, map[string]string{"build": "echo " + id})
	}

	plain := f.page(t, asAdmin, "/")
	assert.Contains(t, plain, `<details class="project" data-project="malta" open>`)
	assert.Contains(t, plain, `<details class="folder" data-folder="malta/deploy"`)
	assert.NotContains(t, plain, `class="workspace`)

	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", Title: "Malta"}))
	body := f.page(t, asAdmin, "/")
	ws := strings.Index(body, `<details class="workspace" data-project="malta" open>`)
	deploy := strings.Index(body, `<details class="project" data-project="malta/deploy" open>`)
	web := strings.Index(body, `<details class="folder" data-folder="malta/deploy/web"`)
	apps := strings.Index(body, `<details class="project" data-project="malta/apps" open>`)
	assert.True(t, ws >= 0 && apps > ws && deploy > apps && web > deploy, "malta holds malta/apps, then malta/deploy with malta/deploy/web")
	assert.NotContains(t, body, `<details class="project" data-project="malta" open>`)
	assert.NotContains(t, body, `data-project="malta/deploy/web"`)
	summary := body[ws : strings.Index(body[ws:], "</summary>")+ws]
	assert.Contains(t, summary, `<svg class="ic chev"><use href="#i-chev"/></svg>`)
	assert.Contains(t, summary, `<a class="section" href="/f/malta" data-folder="malta">`)
	assert.Contains(t, summary, `<small>2</small>`)
	assert.Contains(t, body, `<a class="section" href="/f/malta/deploy" data-folder="malta/deploy">`)
	assert.Contains(t, body, `<a class="section" href="/f/malta/apps" data-folder="malta/apps">`)
	// The sidebar names a folder in a workspace's project below the project.
	assert.Contains(t, body, `<a href="/f/malta/deploy/web"><svg class="ic"><use href="#i-folder"/></svg><span><b>web</b></span></a>`)
	// The workspace closes after its last project.
	closes := ws
	for range 4 {
		closes += strings.Index(body[closes:], "</details>") + len("</details>")
	}
	assert.True(t, closes > web, "the workspace closes after malta/deploy/web")

	// A project at the root lists above the workspaces, whatever its name.
	cat["zeta"] = shellFolder(t, f.root, "zeta", map[string]string{"build": "echo zeta"})
	body = f.page(t, asAdmin, "/")
	zeta := strings.Index(body, `<details class="project" data-project="zeta" open>`)
	ws = strings.Index(body, `<details class="workspace" data-project="malta" open>`)
	assert.True(t, zeta >= 0 && zeta < ws, "zeta lists above malta")

	// The sidebar marks the open page's workspace and project, and lights the folder's row.
	page := f.page(t, asAdmin, "/f/malta/deploy/web")
	assert.Contains(t, page, `<details class="workspace active" data-project="malta" open>`)
	assert.Contains(t, page, `<details class="project active" data-project="malta/deploy" open>`)
	assert.Contains(t, page, `<details class="folder active" data-folder="malta/deploy/web" open>`)
	// The crumbs walk the namespace.
	assert.Contains(t, page, `<div class="crumbs"><a href="/">Dashboard</a><span><a href="/f/malta">malta</a></span><span><a href="/f/malta/deploy">deploy</a></span><span>web</span></div>`)

	// A project of a workspace set not to show leaves with its folders.
	assert.NoError(t, f.opts.Projects.Create(ctx, &model.Project{Folder: "malta/deploy", SidebarHidden: true}))
	hidden := sidebar(f.page(t, asAdmin, "/"))
	assert.Contains(t, hidden, `data-project="malta"`)
	assert.Contains(t, hidden, `data-project="malta/apps"`)
	assert.NotContains(t, hidden, `data-project="malta/deploy"`)
	assert.NotContains(t, hidden, `data-folder="malta/deploy/web"`)
}

// sidebar is the folders part of a page's sidebar.
func sidebar(body string) string {
	start := strings.Index(body, `<nav class="folders">`)
	end := strings.Index(body[start:], "</nav>")
	return body[start : start+end]
}
