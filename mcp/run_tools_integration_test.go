//go:build integration

package mcp_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run of a directory inside a folder reads as folder/dir/job in
// list_runs and get_run.
func TestTools_RunInDir(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	images := filepath.Join(f.root, "site", "images", "scripts")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(images, "where.sh"), []byte("# Where it runs\npwd -P\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, "site", ".git"), 0o755))

	status, raw := f.post(t, f.tokens["admin"], "/api/runs", api.RunRequest{Folder: "site", Dir: "images", Job: "where", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	var view api.RunView
	assert.NoError(t, json.Unmarshal(raw, &view))

	text, _ := f.tool(t, "guest", "list_runs", map[string]any{"folder": "site"})
	assert.Contains(t, text, view.ID+" site/images/where by admin")
	text, failed := f.tool(t, "guest", "get_run", map[string]any{"id": view.ID})
	assert.False(t, failed, text)
	assert.Contains(t, text, "run "+view.ID+": site/images/where by admin")
}
