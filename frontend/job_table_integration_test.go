//go:build integration

package frontend_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A folder page's Files list holds no Action or Summary column: a
// directory that is a folder links to its page, whose Jobs table runs
// its jobs; a file that runs is a row ./<path> of the folder's Jobs
// table, with its last run.
func TestFolder_FilesDashboard(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	web := shellFolder(t, f.root, "infra/web", map[string]string{"build": "true", "default": "true"})
	f.opts.Catalogue.(catalogue)["infra/web"] = web
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "tool.py"), []byte("#!/usr/bin/env python3\n"), 0o755))
	f.run(t, "infra", "./tool.py", asOps, model.StateOK)

	body := f.page(t, asOps, "/f/_/infra")
	assert.Contains(t, body, `<table class="table file-table col-hover">`)
	assert.Contains(t, body, `<thead><tr><th>Name</th><th class="num">Size</th>`)
	assert.NotContains(t, body, "<th>Action</th>")
	assert.NotContains(t, body, "<th>Summary</th>")
	assert.Contains(t, body, `<a href="/f/_/infra/edge">edge</a>`)

	assert.Contains(t, jobsTable(body), `<a class="badge" href="/f/_/infra/web">web/scripts</a>`, "a folder inside with jobs is a row, its jobs its own")
	assert.Contains(t, jobsTable(body), `data-cc="infra/web/build" data-folder="infra/web"`)

	tool := jobRow(t, body, "infra/./tool.py")
	assert.Contains(t, tool, `data-path="tool.py"`)
	assert.Contains(t, tool, `data-run data-folder="infra" data-job="./tool.py"`)
	assert.Contains(t, tool, `class="cc named ok" data-run data-folder="infra" data-job="./tool.py"`)
}

// A directory of a Files list that is no folder, with a job source at
// its top, gets a group of the Jobs table, its heading linked to the
// directory, its jobs run as the folder's in that directory. |> Summarize
// leaves out the interactive ones. The catalogue does not change.
func TestFolder_DirJobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	images := filepath.Join(f.infra.Dir, "images")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(images, "atkins.yml"), []byte("jobs:\n  push:\n    desc: Push one\n    steps:\n      - run: docker compose push\n  default:\n    desc: Every image\n    steps:\n      - run: atkins build\n  shell:\n    interactive: true\n    steps:\n      - run: sh\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "docs"), 0o755))
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	before := len(f.opts.Catalogue.List())

	body := f.page(t, asAdmin, "/f/_/infra")
	table := jobsTable(body)
	assert.Contains(t, table, `<a class="badge" href="/f/_/infra/images">images/atkins.yml</a>`)
	assert.Contains(t, table, `data-run data-folder="infra" data-dir="images" data-job="default" data-title="infra/images / default"`)
	assert.Contains(t, jobRow(t, body, "infra/images/push"), "Push one")
	assert.Contains(t, jobRow(t, body, "infra/images/push"), `<span>Run and summarize push</span>`)
	assert.NotContains(t, table, "Run and summarize shell", "no interactive job to summarize")
	assert.NotContains(t, table, "docs/", "nothing to run")

	// The directory's own page lists them as its own.
	page := f.page(t, asAdmin, "/f/_/infra/images")
	assert.Contains(t, jobsTable(page), `data-dir="images" data-job="push"`)

	assert.Equal(t, before, len(f.opts.Catalogue.List()), "the catalogue stays")
	for _, sidebar := range []string{`data-folder="infra/images"`} {
		assert.NotContains(t, body, sidebar)
	}
}

// jobRow returns the row of the Jobs table that holds a job by its key.
func jobRow(t *testing.T, body, key string) string {
	t.Helper()
	for _, row := range strings.Split(jobsTable(body), "<tr")[1:] {
		if strings.Contains(row, ` data-cc="`+key+`"`) {
			row, _, _ = strings.Cut(row, "</tr>")
			return row
		}
	}
	t.Fatalf("no row of %s", key)
	return ""
}
