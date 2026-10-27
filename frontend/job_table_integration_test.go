//go:build integration

package frontend_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A folder page's Files list holds no Action or Summary column: a
// directory that is a folder links to its page; a file that runs is a
// row of the listing, not of a Jobs table.
func TestFolder_FilesDashboard(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	web := shellFolder(t, f.root, "infra/web", map[string]string{"build": "true", "default": "true"})
	f.opts.Catalogue.(catalogue)["infra/web"] = web
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "tool.py"), []byte("#!/usr/bin/env python3\n"), 0o755))
	f.run(t, "infra", "./tool.py", asOps, model.StateOK)

	body := f.page(t, asOps, "/f/internal/infra")
	assert.Contains(t, body, `<table class="table file-table col-hover">`)
	assert.Contains(t, body, `<thead><tr><th>Name</th><th class="num">Size</th>`)
	assert.NotContains(t, body, "<th>Action</th>")
	assert.NotContains(t, body, "<th>Summary</th>")
	assert.Contains(t, body, `<a href="/f/internal/infra/edge">edge</a>`)
	assert.NotContains(t, body, `data-cc="infra/web/build"`)
}
