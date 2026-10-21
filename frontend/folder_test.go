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
)

func TestFolder(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.run(t, "infra", "plan", "ops", model.StateFailed)

	// Run force stands beside Run only for a job with hours, where the
	// server can refuse a plain run; the window of a failed run forces it.
	f.infra.Job("plan").Hours = "mon-fri 9-17"

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, "The servers")
	// The page lists the folders under it, each linked, with its job count.
	assert.Contains(t, body, `<a href="/f/_/infra/edge">`)
	assert.NotContains(t, body, "Nothing to run was found")
	assert.NotContains(t, body, "data-run-all") // workflows chain jobs instead
	assert.Contains(t, body, `data-run data-folder="infra" data-job="plan"`)
	assert.Contains(t, body, `data-force data-folder="infra" data-job="plan"`)
	assert.Equal(t, 1, strings.Count(body, "data-force"), "only plan has hours")
	assert.Contains(t, body, `<i class="dot failed"></i><span class="state-label">failed`)
	assert.Contains(t, body, ".env")
	assert.Contains(t, body, `class="badge bad">missing`)
	assert.NotContains(t, body, "/f/_/infra/env/.env") // admins edit env files

	viewer := f.page(t, asViewer, "/f/_/infra")
	assert.NotContains(t, viewer, "data-run-all")
	assert.NotContains(t, viewer, "data-force")
	assert.Contains(t, viewer, "History</a>")

	admin := f.page(t, asAdmin, "/f/_/infra")
	assert.Contains(t, admin, `href="/f/_/infra/env/.env"`)
}

func TestFolder_Running(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<i class="dot running"></i><span class="state-label">running</span>`)
	assert.Contains(t, body, `<span class="when">just now</span><span class="who-ran"><span class="avatar sm role-runner" title="ops">OP</span> ops</span></div>`)
	assert.Contains(t, body, `data-attach="`+session.ID()+`"`)
	assert.Contains(t, body, `data-kill="`+session.ID()+`"`)

	viewer := f.page(t, asViewer, "/f/_/infra")
	assert.Contains(t, viewer, `data-attach="`+session.ID()+`"`)
	assert.NotContains(t, viewer, "data-kill")
}

func TestFolder_Missing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/_/nowhere").Status)
	hidden := f.get(t, asLimited, "/f/_/private")
	assert.Equal(t, http.StatusNotFound, hidden.Status)
	assert.Contains(t, hidden.Body, "No folder named private")
}

// The folder page lists what the folder holds, directories first and
// linked, files after, each opening in the preview window; a directory
// with nothing to run opens as a page of its own, down to any depth,
// with .. up a level.
func TestFolder_Tree(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "terraform", "modules"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "README.md"), []byte("infra"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "terraform", "main.tf"), []byte(""), 0o644))

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<a href="/f/_/infra/edge">edge</a>`) // a folder of the catalogue
	assert.Contains(t, body, `<a href="/f/_/infra/terraform">terraform</a>`)
	assert.Contains(t, body, `<a href="/files/infra/README.md" data-file="README.md">README.md</a>`)
	assert.Contains(t, body, `href="/files/infra/README.md?download"`)
	assert.Contains(t, body, `<a href="/f/_" title="Up a level">..</a>`)
	assert.True(t, strings.Index(body, `>terraform</a>`) < strings.Index(body, `>README.md</a>`), "folders before files")

	sub := f.page(t, asOps, "/f/_/infra/terraform")
	assert.Contains(t, sub, "<h2>")
	assert.Contains(t, sub, `<a href="/f/_/infra/terraform/modules">modules</a>`)
	assert.Contains(t, sub, `<a href="/files/infra/terraform/main.tf" data-file="terraform/main.tf">main.tf</a>`)
	assert.Contains(t, sub, `<a href="/f/_/infra">infra</a>`) // the crumb back
	assert.Contains(t, sub, `<a href="/f/_/infra" title="Up a level">..</a>`)
	assert.Contains(t, sub, "Nothing to run")
	deep := f.page(t, asOps, "/f/_/infra/terraform/modules")
	assert.Contains(t, deep, `<a href="/f/_/infra/terraform">terraform</a>`)
	assert.Contains(t, deep, `<a href="/f/_/infra/terraform" title="Up a level">..</a>`)
	edge := f.page(t, asOps, "/f/_/infra/edge")
	assert.Contains(t, edge, `<a href="/f/_/infra" title="Up a level">..</a>`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/_/infra/nowhere").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/_/infra/terraform/README.md").Status) // a file is not a page
	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/f/_/private/anything").Status)
}

// The page lists dot entries; the sealed ones, .git and every dot
// file, list muted, with no link and no Run, and neither the pages nor
// /files/ open them, anything in .git, or a symlink to either. A dot
// directory opens, and what is in it but its dot files. A symlink out
// of the checkout opens nothing either.
func TestFolder_Sealed(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := f.infra.Dir
	for name, content := range map[string]string{".git/config": "[core]\n", ".env": "A=1\n", ".env.local": "B=2\n", "web/.htaccess": "deny\n", ".claude/settings.json": "{}\n", ".claude/.secret": "s\n", ".gitignore": "x\n"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	assert.NoError(t, os.WriteFile(filepath.Join(f.root, "outside.txt"), []byte("outside\n"), 0o644))
	assert.NoError(t, os.Symlink(filepath.Join(f.root, "outside.txt"), filepath.Join(dir, "escape.txt")))
	assert.NoError(t, os.Symlink(f.root, filepath.Join(dir, "rootlink")))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".hook"), []byte("#!/bin/sh\necho hook\n"), 0o755))
	assert.NoError(t, os.Symlink(".env", filepath.Join(dir, "envlink")))

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<tr class="sealed" aria-disabled="true" title=".git is listed, but task-ui does not open it">`)
	assert.Contains(t, body, `<span>.git</span>`)
	assert.Contains(t, body, `<span>.env</span>`)
	assert.Contains(t, body, `<span>.env.local</span>`)
	assert.NotContains(t, body, `href="/f/_/infra/.git"`)
	assert.NotContains(t, body, `/files/infra/.env`)
	assert.Contains(t, body, `<a href="/f/_/infra/.claude">.claude</a>`)
	assert.Contains(t, body, `<span>.gitignore</span>`)
	assert.Contains(t, body, `<span>.hook</span>`)
	assert.Contains(t, body, `<span>envlink</span>`)
	assert.NotContains(t, body, `data-job="./.hook"`)
	assert.NotContains(t, body, `/files/infra/envlink`)
	claude := f.page(t, asOps, "/f/_/infra/.claude")
	assert.Contains(t, claude, `data-file=".claude/settings.json"`)
	assert.Contains(t, claude, `<span>.secret</span>`)
	assert.Equal(t, http.StatusOK, f.get(t, asOps, "/files/infra/.claude/settings.json").Status)
	assert.Contains(t, body, `<span>escape.txt</span>`)
	assert.NotContains(t, body, `/files/infra/escape.txt`)
	assert.Contains(t, f.page(t, asOps, "/f/_/infra/web"), `<span>.htaccess</span>`)

	for _, path := range []string{"/f/_/infra/.git", "/f/_/infra/rootlink", "/f/_/infra/web/.htaccess", "/f/_/infra/.hook", "/f/_/infra/./.hook"} {
		assert.Equal(t, http.StatusNotFound, f.get(t, asOps, path).Status, path)
	}
	for _, path := range []string{
		"/files/infra/.git/config", "/files/infra/.env", "/files/infra/.env.local", "/files/infra/web/.htaccess",
		"/files/infra/escape.txt", "/files/infra/rootlink/outside.txt", "/files/infra/../outside.txt", "/files/infra/web/../.env",
		"/files/infra/.gitignore", "/files/infra/.hook", "/files/infra/envlink", "/files/infra/.claude/.secret",
	} {
		assert.Equal(t, http.StatusNotFound, f.get(t, asOps, path).Status, path)
	}
}

// A file that runs carries #! and exec, and has Run for those the
// folder lets run, which starts the job ./<path>; its history opens by
// its path.
func TestFolder_Executable(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "bin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "bin", "hello"), []byte("#!/bin/sh\necho hello\n"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "tool.py"), []byte("#!/usr/bin/env python3\n"), 0o644))

	ops := f.page(t, asOps, "/f/_/infra/bin")
	assert.Contains(t, ops, `<i class="shebang" aria-hidden="true">#!</i> <a href="/files/infra/bin/hello" data-file="bin/hello">hello</a> <span class="badge exec"`)
	assert.Contains(t, ops, `data-run data-folder="infra" data-job="./bin/hello" data-title="infra / bin/hello"`)
	assert.Contains(t, f.page(t, asOps, "/f/_/infra"), `data-job="./tool.py"`)

	viewer := f.page(t, asViewer, "/f/_/infra/bin")
	assert.Contains(t, viewer, `<span class="badge exec"`)
	assert.NotContains(t, viewer, `data-job="./bin/hello"`)

	job := f.get(t, asOps, "/f/_/infra/bin/hello")
	assert.Equal(t, http.StatusOK, job.Status)
	assert.Contains(t, job.Body, "<h2>bin/hello")
	assert.Equal(t, http.StatusOK, f.get(t, asOps, "/f/_/infra/./bin/hello").Status)
}

// The sidebar heads each top level folder with its name and its jobs,
// marks the folder of the page, a menu button drives the drawer, and the
// jobs table folds into cards.
func TestFolder_SidebarAndCards(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<a class="section active" href="/f/_/infra" data-folder="infra"`)
	// The sidebar runs a job in place and attaches to the one running.
	assert.Contains(t, body, `<a class="section" href="/f/_/private" data-folder="private"`)
	assert.Contains(t, body, `<div class="job" data-job-key="infra/plan">`)
	assert.Contains(t, body, `class="play" data-run data-folder="infra" data-job="plan"`)
	assert.Contains(t, body, `<div class="job" data-job-key="infra/slow">
	<i class="shebang running" aria-hidden="true">#!</i>`)
	// Every job carries #!, the card's in the colour of its last run.
	assert.Contains(t, body, `<div class="job" data-job-key="infra/plan">
	<i class="shebang" aria-hidden="true">#!</i>`)
	assert.Contains(t, body, `class="play" data-attach="`+session.ID()+`"`)
	assert.NotContains(t, body, `class="play" data-run data-folder="private"`) // ops may not run private
	assert.Contains(t, body, `id="menu-toggle"`)
	assert.Contains(t, body, `id="menu-backdrop"`)
	assert.Contains(t, body, `<div class="jobcards">`)
	assert.Contains(t, body, `<article class="jobcard">`)
	assert.Contains(t, body, `<article class="jobcard running">`)
	assert.Contains(t, body, "<th>Hours</th><td><span class=\"faint\">always</span></td>")

	job := f.page(t, asOps, "/f/_/infra/plan")
	assert.Contains(t, job, `<div class="job active" data-job-key="infra/plan">`)

	viewer := f.page(t, asViewer, "/f/_/infra")
	assert.NotContains(t, viewer, `class="play" data-run`)
	assert.Contains(t, viewer, `class="play" data-attach="`+session.ID()+`"`)
}

func TestFolder_Hours(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.infra.Hours = "mon-fri 9-17\nfri 9-13"
	f.infra.Job("plan").Hours = "daily 0-24"

	// The fixture's clock is a Friday night.
	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, "mon-fri 9-17, fri 9-13 <span class=\"badge bad\">out of hours now</span>")

	job := f.page(t, asOps, "/f/_/infra/plan")
	assert.Contains(t, job, "daily 0-24 <span class=\"badge ok\">open now</span>")
	assert.Contains(t, job, "the folder runs mon-fri 9-17, fri 9-13")

	slow := f.page(t, asOps, "/f/_/infra/slow")
	assert.Contains(t, slow, "out of hours now")
}

// A nested folder has a path for an ID: its page, its jobs and its env
// files resolve under it, and the sidebar lists it under the heading of
// its top level folder, by its path below it.
func TestFolder_Nested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/f/_/infra/edge")
	assert.Contains(t, body, `<details class="folder active" data-folder="infra/edge" open>`)
	assert.Contains(t, body, `<a class="section" href="/f/_/infra" data-folder="infra"`)
	assert.Contains(t, body, `<a href="/f/_/infra/edge"><svg class="ic"><use href="#i-folder-open"/></svg><span><b>edge</b></span></a>`)
	assert.Contains(t, body, `<a href="/f/_/infra">infra</a>`)
	assert.Contains(t, body, `class="play" data-run data-folder="infra/edge" data-job="deploy"`)

	job := f.page(t, asAdmin, "/f/_/infra/edge/deploy")
	assert.Contains(t, job, "<h2>deploy")
	assert.Contains(t, job, `<span><a href="/f/_/infra/edge">edge</a></span><span>deploy</span>`)

	env := f.page(t, asAdmin, "/f/_/infra/edge/env/.env")
	assert.Contains(t, env, `data-api="/api/env/infra/edge/.env"`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/f/_/infra/edge/nothing").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/f/_/infra/edge/deploy/more").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/f/_/infra/nope/deploy").Status)
	// The limited user sees infra; the folder inside it stays hidden.
	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/f/_/infra/edge").Status)
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/settings/projects/infra/edge").Status)
}

// A folder's path starts with its workspace: a named one by its folder,
// _ for a project at the root. A path that starts with neither moves to
// /f/_/ for good, its query kept; internal stays put, and a named
// workspace is not reached through _.
func TestFolder_Namespace(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	cat := f.opts.Catalogue.(catalogue)
	for _, id := range []string{"acme", "acme/api", model.Internal} {
		cat[id] = shellFolder(t, f.root, id, map[string]string{"build": "echo " + id})
	}
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))

	moved := f.get(t, asOps, "/f/infra/plan?notice=Saved#x")
	assert.Equal(t, http.StatusMovedPermanently, moved.Status)
	assert.Equal(t, "/f/_/infra/plan?notice=Saved", moved.Location)
	assert.Equal(t, "/f/_/infra", f.get(t, asOps, "/f/infra").Location)

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<a class="section active" href="/f/_/infra" data-folder="infra">`)
	assert.Contains(t, body, `<a href="/f/_/infra/edge">`)
	assert.Contains(t, body, `<a class="name" href="/f/_/infra/plan">`)
	assert.Contains(t, body, `<a class="section" href="/f/acme" data-folder="acme">`)
	assert.Contains(t, body, `<a class="section" href="/f/acme/api" data-folder="acme/api">`)
	assert.Contains(t, body, `data-workspaces='{&#34;acme&#34;:0}'`)
	assert.Contains(t, body, `<div class="crumbs"><a href="/">Dashboard</a><span><a href="/f/_">_</a></span><span>infra</span></div>`)
	assert.NotContains(t, sidebar(body), `href="/f/infra`)
	assert.Contains(t, f.page(t, asOps, "/f/_/infra/edge"), `<span><a href="/f/_/infra">infra</a></span><span>edge</span>`)

	assert.Contains(t, f.page(t, asOps, "/f/acme/api"), `<span><a href="/f/acme">acme</a></span><span>api</span>`)
	for _, path := range []string{"/f/_/acme", "/f/_/acme/api", "/f/_/internal", "/f/_/_", "/f/_/nothing"} {
		assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, path).Status, path)
	}
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/f/internal").Status)

	// /f/_ is the default workspace: the projects at the root a user sees.
	def := f.page(t, asLimited, "/f/_")
	assert.Contains(t, def, "Default")
	assert.Contains(t, def, `<a href="/f/_/infra">`)
	assert.NotContains(t, def, `<a href="/f/_/private">`)
	assert.NotContains(t, def, `href="/f/acme`)
	assert.NotContains(t, def, "the root</small>")
	admin := f.page(t, asAdmin, "/f/_")
	assert.Contains(t, admin, `<a href="/f/_/private">`)
	assert.Contains(t, admin, "the root</small>")
	assert.Contains(t, admin, `<a href="/f/acme">acme</a>`)
}

// A job card names the job once, with its kind beside it; it numbers
// the lines it runs, shows the first three and keeps the rest behind
// "show N more"; it drops a description that repeats the name or the
// first step.
func TestFolder_JobCard(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	plan := f.infra.Job("plan")
	plan.Description = "Show the plan"
	plan.Steps = []string{"task setup:checkout", "go build ./...", "go test ./...", "docker build .", "docker push"}
	plan.DependsOn = []string{"lint"}
	f.infra.Job("slow").Description = "slow"
	f.infra.Job("many").Description = "seq 1 200"
	f.infra.Job("many").Steps = []string{"seq 1 200"}

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<a class="name" href="/f/_/infra/plan">plan</a>
				<span class="badge">script</span>`)
	assert.Contains(t, body, `<p class="desc">Show the plan</p>`)
	assert.Contains(t, body, `<div class="code"><div class="comment"># after lint</div>
				<ol><li>task setup:checkout</li><li>go build ./...</li><li>go test ./...</li></ol>
				<details>
					<summary><span class="more">show 2 more</span><span class="less">show less</span></summary>
					<ol start="4"><li>docker build .</li><li>docker push</li></ol>
				</details>`)
	// The name and the first step are not said twice.
	slow := card(body, "slow")
	assert.Equal(t, 1, strings.Count(slow, ">slow<"))
	assert.Contains(t, slow, `<p class="desc"></p>`)
	many := card(body, "many")
	assert.Equal(t, 1, strings.Count(many, "seq 1 200"))
	assert.Contains(t, many, `<p class="desc"></p>`)
	// Three lines or fewer have no control; a job without steps shows its command.
	assert.Equal(t, 1, strings.Count(body, "<details>"))
	assert.Contains(t, body, `<ol><li>sh -c sleep 30</li></ol>`)
	// One meta row: the state, then when, then who.
	assert.Contains(t, body, `<div class="meta"><i class="dot pending"></i><span class="state-label faint">never run</span></div>`)
}

// card cuts the job card of a job out of a folder page.
func card(body, job string) string {
	_, rest, _ := strings.Cut(body, `<a class="name" href="/f/_/infra/`+job+`">`)
	rest, _, _ = strings.Cut(rest, "</article>")
	return `<a class="name">` + rest
}

// A directory browsed inside a folder lists its own targets as job
// cards, with Run naming the directory and the job's page under it; a
// run there shows on its card and in that page's history, not in the
// history of the folder's job of the same name.
func TestFolder_BrowseJobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	images := filepath.Join(f.infra.Dir, "images")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(images, "atkins.yml"), []byte("jobs:\n  default:\n    desc: Every image\n    steps:\n      - run: atkins build\n  build:\n    desc: Build one\n    steps:\n      - run: docker compose build\n  push:\n    desc: Push one\n    steps:\n      - run: docker compose push\n"), 0o644))
	run := f.run(t, "infra", "build", "ops", model.StateFailed)
	run.Dir = "images"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), run))

	body := f.page(t, asOps, "/f/_/infra/images")
	assert.Contains(t, body, "Its jobs run in infra/images/")
	assert.NotContains(t, body, "Nothing to run here")
	for _, job := range []string{"default", "build", "push"} {
		assert.Contains(t, body, `<a class="name" href="/f/_/infra/images/`+job+`">`+job+`</a>`)
		assert.Contains(t, body, `data-folder="infra" data-dir="images" data-job="`+job+`" data-title="infra/images / `+job+`"`)
	}
	assert.Contains(t, body, `<i class="dot failed"></i>`)
	assert.Contains(t, body, "atkins -f atkins.yml")
	viewer := f.page(t, asViewer, "/f/_/infra/images")
	assert.Contains(t, viewer, `<a class="btn btn-sm" href="/f/_/infra/images/build">History</a>`)
	assert.NotContains(t, viewer, `data-dir="images"`)

	page := f.page(t, asOps, "/f/_/infra/images/build")
	assert.Contains(t, page, "infra/images / build")
	assert.Contains(t, page, `data-attach="`+run.ID+`"`)
	assert.Contains(t, page, "the job runs here")
	assert.NotContains(t, page, "<h3>Scheduled")
	own := f.run(t, "infra", "plan", "ops", model.StateOK)
	inDir := f.run(t, "infra", "plan", "ops", model.StateOK)
	inDir.Dir = "images"
	assert.NoError(t, f.opts.Runs.Update(context.Background(), inDir))
	_, plan, _ := strings.Cut(f.page(t, asOps, "/f/_/infra/plan"), "<h3>History")
	plan, _, _ = strings.Cut(plan, "</section>")
	assert.Contains(t, plan, `data-attach="`+own.ID+`"`)
	assert.NotContains(t, plan, `data-attach="`+inDir.ID+`"`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/f/_/infra/images/nothing").Status)
}
