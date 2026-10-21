package api_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// POST /api/prompts runs internal/claude-prompt in a folder, or in a
// directory inside one, with TASK_UI_PROMPT_DIR naming it: admins only,
// one prompt per directory at a time, two directories at once.
func TestPrompts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptFolder(t, f)
	out := filepath.Join(f.root, model.Internal, "env.out")
	infra := filepath.Join(f.root, "infra")

	for _, user := range []string{asOps, asViewer, asLimited} {
		res := f.call(t, user, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello"})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}

	res := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: " what runs here ", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, model.Internal, res.str("folder"))
	assert.Equal(t, "claude-prompt", res.str("job"))
	assert.Equal(t, "infra", res.str("dir"))
	assert.Equal(t, model.StateOK, f.finished(t, res.str("id")).State)
	env, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Equal(t, "TASK_UI_PROMPT=what runs here\nTASK_UI_PROMPT_DIR="+infra+"\npwd="+infra+"\n", string(env))

	// A prompt holds its directory: another there answers 409 with it,
	// one in another directory runs.
	hold := func(rel string) string {
		dir := filepath.Join(infra, rel)
		assert.NoError(t, os.MkdirAll(dir, 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "hold"), nil, 0o644))
		return dir
	}
	images, docs := hold("images"), hold("docs")
	first := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "images", Prompt: "build"})
	assert.Equal(t, http.StatusCreated, first.Status, string(first.Raw))
	assert.Equal(t, "infra/images", first.str("dir"))
	busy := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "images/", Prompt: "again"})
	assert.Equal(t, http.StatusConflict, busy.Status, string(busy.Raw))
	assert.Equal(t, first.str("id"), busy.str("run", "id"))
	second := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "docs", Prompt: "read"})
	assert.Equal(t, http.StatusCreated, second.Status, string(second.Raw))
	assert.Equal(t, model.StateRunning, f.call(t, asAdmin, http.MethodGet, "/runs/"+first.str("id"), nil).str("state"))
	for _, dir := range []string{images, docs} {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "release"), nil, 0o644))
	}
	assert.Equal(t, model.StateOK, f.finished(t, first.str("id")).State)
	assert.Equal(t, model.StateOK, f.finished(t, second.str("id")).State)

	for _, c := range []struct {
		name string
		req  api.PromptRequest
		want int
	}{
		{"empty", api.PromptRequest{Folder: "infra", Prompt: "  "}, http.StatusBadRequest},
		{"no folder", api.PromptRequest{Folder: "nowhere", Prompt: "x"}, http.StatusNotFound},
		{"internal", api.PromptRequest{Folder: model.Internal, Prompt: "x"}, http.StatusNotFound},
		{"up and out", api.PromptRequest{Folder: "infra", Dir: "../private", Prompt: "x"}, http.StatusNotFound},
		{"missing", api.PromptRequest{Folder: "infra", Dir: "nothing", Prompt: "x"}, http.StatusNotFound},
		{"a file", api.PromptRequest{Folder: "infra", Dir: "images/hold", Prompt: "x"}, http.StatusNotFound},
	} {
		res := f.call(t, asAdmin, http.MethodPost, "/prompts", c.req)
		assert.Equal(t, c.want, res.Status, c.name+": "+strings.TrimSpace(string(res.Raw)))
	}
}
