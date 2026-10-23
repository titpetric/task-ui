//go:build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// checkoutFiles writes into infra a text file, markdown, an image, a
// binary, a large file, a script, the sealed entries and symlinks in and
// out of it.
func checkoutFiles(t *testing.T, f *fixture) string {
	t.Helper()
	dir := f.opts.Catalogue.Find("infra").Dir
	files := map[string]string{
		"notes.txt":            "a <b> c\n",
		"docs/README.md":       "# Infra\n\n![plot](plot.png)\n",
		"docs/plot.png":        "\x89PNG\r\n",
		"blob.bin":             "a\x00b",
		"big.log":              strings.Repeat("é", 1<<20),
		".git/config":          "[core]\n",
		".env":                 "TOKEN=secret\n",
		"sub/.env.local":       "TOKEN=local\n",
		"web/.htpasswd":        "u:x\n",
		".claude/CLAUDE.md":    "# hello\n",
		".claude/.secret":      "TOKEN=claude\n",
		".atkins/skills/x.yml": "x: 1\n",
		".gitignore":           "*.log\n",
	}
	for name, content := range files {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "hello"), []byte("#!/bin/sh\necho hello from $(basename $0)\n"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".hook"), []byte("#!/bin/sh\necho hook\n"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", ".run"), []byte("#!/bin/sh\necho run\n"), 0o755))
	assert.NoError(t, os.Symlink(".env", filepath.Join(dir, "envlink")))
	assert.NoError(t, os.Symlink(".hook", filepath.Join(dir, "hook")))
	assert.NoError(t, os.WriteFile(filepath.Join(f.root, "outside.txt"), []byte("outside\n"), 0o644))
	assert.NoError(t, os.Symlink(filepath.Join(f.root, "outside.txt"), filepath.Join(dir, "escape.txt")))
	assert.NoError(t, os.Symlink(".git/config", filepath.Join(dir, "gitconfig")))
	return dir
}

// A file of a checkout opens for whoever sees the folder: text coloured,
// markdown rendered, an image by its kind, a binary as a file, a large
// text cut. Nothing outside the checkout or sealed opens.
func TestFile(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	checkoutFiles(t, f)

	text := f.call(t, asViewer, http.MethodGet, "/files/infra/notes.txt", nil)
	assert.Equal(t, http.StatusOK, text.Status, string(text.Raw))
	assert.Equal(t, "text", text.str("kind"))
	assert.Equal(t, "a <b> c\n", text.str("text"))
	assert.Contains(t, text.str("html"), "a &lt;b&gt; c")

	md := f.call(t, asViewer, http.MethodGet, "/files/infra/docs/README.md", nil)
	assert.Equal(t, "markdown", md.str("kind"))
	assert.Contains(t, md.str("html"), `<img src="/files/infra/docs/plot.png" alt="plot" />`)
	assert.Equal(t, "image", f.call(t, asViewer, http.MethodGet, "/files/infra/docs/plot.png", nil).str("kind"))
	assert.Equal(t, "file", f.call(t, asViewer, http.MethodGet, "/files/infra/blob.bin", nil).str("kind"))
	assert.Equal(t, "markdown", f.call(t, asViewer, http.MethodGet, "/files/infra/.claude/CLAUDE.md", nil).str("kind")) // a dot directory opens
	assert.Equal(t, "text", f.call(t, asViewer, http.MethodGet, "/files/infra/.atkins/skills/x.yml", nil).str("kind"))

	big := f.call(t, asViewer, http.MethodGet, "/files/infra/big.log", nil)
	assert.Equal(t, "text", big.str("kind"))
	assert.Equal(t, true, big.Body["truncated"])
	assert.Equal(t, float64(2<<20), big.Body["byte_size"])
	assert.True(t, len(big.str("text")) <= 1<<20)

	for _, path := range []string{
		"infra/../private/x", "infra/docs/../notes.txt", "infra/escape.txt", "infra/gitconfig",
		"infra/.git", "infra/.git/config", "infra/.env", "infra/sub/.env.local", "infra/web/.htpasswd",
		"infra/docs", "infra/missing.txt", "infra", "nowhere/notes.txt",
		"infra/.gitignore", "infra/.hook", "infra/.claude/.secret", "infra/.claude/.run", "infra/envlink", "infra/hook",
	} {
		assert.Equal(t, http.StatusNotFound, f.call(t, asViewer, http.MethodGet, "/files/"+path, nil).Status, path)
	}
	// Nor in a folder the user does not see, or task-ui's own.
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/files/private/x", nil).Status)
}

// A file in a directory of the checkout runs from the top of it, by
// its path: bin/where, this test binary, runs in the folder with
// argv[0] ./bin/where.
func TestFile_RunFromRoot(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir, err := filepath.EvalSymlinks(checkoutFiles(t, f))
	assert.NoError(t, err)
	self, err := os.Executable()
	assert.NoError(t, err)
	assert.NoError(t, os.Link(self, filepath.Join(dir, "bin", "where")))

	res := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "./bin/where", Cols: 200, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	var view api.RunView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, model.StateOK, f.finished(t, view.ID).State)
	out := string(f.call(t, asOps, http.MethodGet, "/runs/"+view.ID+"/frame.txt", nil).Raw)
	assert.Contains(t, out, "cwd="+dir+"\n")
	assert.Contains(t, out, "argv0=./bin/where")
}

// The view decodes into its type.
func TestFileView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	checkoutFiles(t, f)
	res := f.call(t, asOps, http.MethodGet, "/files/infra/notes.txt", nil)
	var view api.FileView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, api.FileView{Folder: "infra", Path: "notes.txt", Kind: "text", ContentType: "text/plain; charset=utf-8", ByteSize: 8, Text: "a <b> c\n", HTML: "a &lt;b&gt; c\n"}, view)
}

// A file that runs starts as the job ./<path> of its folder, for those
// the folder lets run; the API refuses a viewer and a user outside the
// allow list, and a file that does not run is no job.
func TestFile_Run(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	checkoutFiles(t, f)

	res := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "./bin/hello", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	var view api.RunView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "./bin/hello", view.Job)
	assert.True(t, view.CanRun)
	run := f.finished(t, view.ID)
	assert.Equal(t, model.StateOK, run.State)
	out := f.call(t, asOps, http.MethodGet, "/runs/"+view.ID+"/frame.txt", nil)
	assert.Contains(t, string(out.Raw), "hello from hello")

	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "./bin/hello"}).Status)
	for _, job := range []string{"./notes.txt", "./bin", "./../infra/bin/hello", "./escape.txt", "./.git/config", "./.hook", "./hook", "./.claude/.run"} {
		assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: job}).Status, job)
	}

	private := f.opts.Catalogue.Find("private").Dir
	assert.NoError(t, os.WriteFile(filepath.Join(private, "go.sh"), []byte("#!/bin/sh\necho private\n"), 0o755))
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "private", Job: "./go.sh"}).Status)
}
