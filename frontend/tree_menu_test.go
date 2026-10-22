package frontend_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A folder page's Files list is a dashboard of what the folder holds:
// no Folders table, a directory that is a folder links to its page and
// holds Run in Jobs, a split button: the default or the only job, with
// a dot of how it last ran, and a menu of the jobs beside it, default
// first and the rest by name. A file that runs and is no job source has
// its plain Run. The dots follow the runs.
func TestFolder_FilesDashboard(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	web := shellFolder(t, f.root, "infra/web", map[string]string{"build": "true", "default": "true", "Lint": "true", "test": "true", "zap": "true"})
	folder.DefaultFirst(web.Jobs)
	web.Job("zap").Allow = []string{model.RoleAdmin}
	f.opts.Catalogue.(catalogue)["infra/web"] = web
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "tool.py"), []byte("#!/usr/bin/env python3\n"), 0o755))
	f.run(t, "infra/edge", "deploy", asOps, model.StateOK)
	f.run(t, "infra/web", "build", asOps, model.StateFailed)

	body := f.page(t, asOps, "/f/_/infra")
	assert.NotContains(t, body, "<h3>Folders</h3>")
	assert.Contains(t, body, `<table class="table file-table col-hover">`)
	assert.Contains(t, body, `<th>Name</th><th>Jobs</th>`)
	assert.Contains(t, body, `<a href="/f/_/infra/edge">edge</a>`)

	edge := treeRow(t, body, "edge")
	assert.Contains(t, edge, `<span class="runsplit" data-key="infra/edge/deploy"><button type="button" class="btn btn-sm btn-primary" data-run data-folder="infra/edge" data-job="deploy" data-title="infra/edge / deploy" title="Run deploy"><svg class="ic"><use href="#i-play"/></svg>Run<i class="dot ok"></i></button>`)
	assert.NotContains(t, edge, "run-menu", "one job runs with no menu")

	webRow := treeRow(t, body, "web")
	assert.Contains(t, webRow, `data-key="infra/web/default"`)
	assert.Contains(t, webRow, `data-job="default" data-title="infra/web / default" title="Run default"><svg class="ic"><use href="#i-play"/></svg>Run<i class="dot pending"></i></button>`)
	assert.Contains(t, webRow, `<details class="cc-more run-menu">`)
	assert.Contains(t, webRow, `data-job="build" data-title="infra/web / build"><td class="glyph failed" title="Last run: failed">`)
	assert.Contains(t, webRow, `<td class="name">build</td><td class="desc">Run the build target</td>`)
	order := []int{}
	for _, job := range []string{"default", "build", "Lint", "test", "zap"} {
		order = append(order, strings.Index(webRow, `<td class="name">`+job+`</td>`))
	}
	for i := 1; i < len(order); i++ {
		assert.True(t, order[0] > 0 && order[i-1] < order[i], "default first, the rest by name, case aside")
	}
	assert.Contains(t, webRow, `<tr class="idle" aria-disabled="true" title="No permission to run zap"><td class="glyph"><i class="shebang" aria-hidden="true">#!</i></td><td class="name">zap</td>`)
	assert.Contains(t, body, `<td class="jobs"><div class="actions"><button type="button" class="btn btn-primary btn-sm" data-run data-folder="infra" data-job="./tool.py"`)

	viewer := f.page(t, asViewer, "/f/_/infra")
	assert.Contains(t, treeRow(t, viewer, "edge"), `<a class="btn btn-sm" href="/f/_/infra/edge/deploy" title="No permission to run deploy"><i class="dot ok"></i>History</a>`)

	js := f.get(t, asOps, "/static/task-ui.js").Body
	assert.Contains(t, js, `if (document.querySelector(".runsplit[data-key]") && window.EventSource) {`)
	assert.Contains(t, js, `.cc-menu [data-run], .cc-menu [data-attach]`)
}

// A directory of a Files list that is no folder, with a job source at
// its top, has Run and |>summarize too, its jobs run as the folder's in
// that directory; the menu is a table of the targets, the description
// or "Run the <name> target", and |>summarize leaves out the interactive
// ones, and runs the one remaining directly. A file that is a job source
// has the same menu. The catalogue does not change.
func TestFolder_RunMenus(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	images := filepath.Join(f.infra.Dir, "images")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(images, "atkins.yml"), []byte("jobs:\n  push:\n    desc: Push one\n    steps:\n      - run: docker compose push\n  default:\n    desc: Every image\n    steps:\n      - run: atkins build\n  Build:\n    steps:\n      - run: docker compose build\n  shell:\n    interactive: true\n    steps:\n      - run: sh\n"), 0o644))
	solo := filepath.Join(f.infra.Dir, "solo")
	assert.NoError(t, os.MkdirAll(solo, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(solo, "Makefile"), []byte("check:\n\ttrue\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "docs"), 0o755))
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	before := len(f.opts.Catalogue.List())

	body := f.page(t, asAdmin, "/f/_/infra")
	row := treeRow(t, body, "images")
	assert.Contains(t, row, `<span class="runsplit" data-key="infra/images/default"><button type="button" class="btn btn-sm btn-primary" data-run data-folder="infra" data-dir="images" data-job="default" data-title="infra/images / default" title="Run default">`)
	assert.Contains(t, row, `<td class="name">Build</td><td class="desc">Run the Build target</td>`)
	assert.Contains(t, row, `<td class="name">push</td><td class="desc">Push one</td>`)
	run, sum, _ := strings.Cut(row, "data-summarize")
	assert.True(t, strings.Index(run, `>default</td>`) < strings.Index(run, `>Build</td>`) && strings.Index(run, `>Build</td>`) < strings.Index(run, `>push</td>`) && strings.Index(run, `>push</td>`) < strings.Index(run, `>shell</td>`), "default first, the rest by name")
	assert.Contains(t, sum, `>push</td>`)
	assert.NotContains(t, sum, `>shell</td>`, "no interactive job to summarize")
	assert.Contains(t, row, `<tr role="menuitem" tabindex="0" data-run data-summarize data-folder="infra" data-dir="images" data-job="Build"`)

	soloRow := treeRow(t, body, "solo")
	assert.Contains(t, soloRow, `data-run data-folder="infra" data-dir="solo" data-job="check"`)
	assert.Contains(t, soloRow, `data-run data-summarize data-folder="infra" data-dir="solo" data-job="check"`)
	assert.NotContains(t, soloRow, "run-menu", "one job runs directly")
	assert.NotContains(t, treeRow(t, body, "docs"), "runsplit", "nothing to run")

	// The atkins.yml of the directory's page is a job source: its row has
	// the same menu.
	page := f.page(t, asAdmin, "/f/_/infra/images")
	file := treeRow(t, page, "atkins.yml")
	assert.Contains(t, file, `data-key="infra/images/default"`)
	assert.Contains(t, file, `<td class="name">Build</td><td class="desc">Run the Build target</td>`)

	// Summarize with one target left runs it directly.
	assert.NoError(t, os.WriteFile(filepath.Join(solo, "atkins.yml"), []byte("jobs:\n  build:\n    steps:\n      - run: true\n  sh:\n    interactive: true\n    steps:\n      - run: sh\n"), 0o644))
	soloPage := f.page(t, asAdmin, "/f/_/infra/solo")
	soloFile := treeRow(t, soloPage, "atkins.yml")
	_, soloSum, _ := strings.Cut(soloFile, "data-summarize")
	assert.Contains(t, soloFile, "run-menu", "Run lists both")
	assert.NotContains(t, soloSum, "run-menu", "summarize runs build directly")
	assert.Contains(t, soloFile, `data-run data-summarize data-folder="infra" data-dir="solo" data-job="build"`)

	assert.Equal(t, before, len(f.opts.Catalogue.List()), "the catalogue stays")
	for _, sidebar := range []string{`data-folder="infra/images"`, `data-folder="infra/solo"`} {
		assert.NotContains(t, body, sidebar)
	}
}

// treeRow returns the row of a Files list that names an entry.
func treeRow(t *testing.T, body, name string) string {
	t.Helper()
	for _, row := range strings.Split(body, "\n\t\t<tr class=")[1:] {
		if strings.Contains(row, ">"+name+"</a>") || strings.Contains(row, ">"+name+"</span>") {
			return row
		}
	}
	t.Fatalf("no row of %s", name)
	return ""
}
