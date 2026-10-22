package api_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// POST /api/prompts puts a prompt to Claude in a folder, or in a
// directory inside one: a note, which internal/claude-prompt answers in
// that directory, with TASK_UI_PROMPT_DIR naming it, as a reply. Admins
// only.
func TestPrompts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptFolder(t, f)
	out := filepath.Join(f.root, model.Internal, "env.out")
	infra := filepath.Join(f.root, "infra")

	for _, user := range []string{asOps, asViewer, asLimited} {
		res := f.call(t, user, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello"})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
		res = f.call(t, user, http.MethodGet, "/prompts?folder=infra", nil)
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}

	res := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: " what runs here ", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "what runs here", res.str("body"))
	assert.Equal(t, asAdmin, res.str("user_name"))
	assert.Equal(t, model.NoteOpen, res.str("state"))
	assert.Equal(t, "infra", res.str("dir"))
	assert.NotEqual(t, "", res.str("workspace_id"))
	asked := f.answered(t, res.str("id"))
	assert.Equal(t, model.NoteDone, asked.State)
	run := f.finished(t, asked.RunID)
	assert.Equal(t, model.Internal, run.Folder)
	assert.Equal(t, "claude-prompt", run.Job)
	assert.Equal(t, "infra", run.Dir)
	assert.Equal(t, asAdmin, run.UserName)
	env, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Equal(t, "TASK_UI_PROMPT=what runs here\nTASK_UI_PROMPT_DIR="+infra+"\npwd="+infra+"\n", string(env))

	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?folder=infra", nil)
	assert.Equal(t, http.StatusOK, threads.Status, string(threads.Raw))
	assert.Len(t, threads.list("notes"), 1)
	replies := func() []any {
		notes := f.call(t, asAdmin, http.MethodGet, "/prompts?folder=infra", nil).list("notes")
		list, _ := notes[0].(map[string]any)["replies"].([]any)
		return list
	}
	answer := replies()[0].(map[string]any)
	assert.Equal(t, "claude", answer["user_name"])
	assert.Equal(t, "Answer to: what runs here", answer["body"])
	assert.Equal(t, asked.RunID, answer["run_id"])
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes", nil).list("notes"), 0, "the board lists no prompt")

	// A reply asks again, with the thread.
	again := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Parent: asked.ID, Prompt: "and then"})
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))
	assert.Equal(t, asked.ID, again.str("parent_id"))
	assert.Equal(t, "infra", again.str("dir"))
	assert.Eventually(t, func() bool { return len(replies()) == 3 }, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, model.NoteDone, f.answered(t, asked.ID).State)
	last := replies()[2].(map[string]any)
	assert.Equal(t, "Answer to: admin:\nwhat runs here\n\nclaude:\nAnswer to: what runs here\n\nadmin:\nand then", last["body"])

	// A prompt holds its directory: another there waits for it, one in
	// another directory runs at once.
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
	queued := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "images/", Prompt: "again"})
	assert.Equal(t, http.StatusCreated, queued.Status, string(queued.Raw))
	second := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "docs", Prompt: "read"})
	assert.Equal(t, http.StatusCreated, second.Status, string(second.Raw))
	working := func(id string) bool {
		note, err := f.opts.Board.Notes.Get(t.Context(), id)
		return err == nil && note.State == model.NoteWorking
	}
	assert.Eventually(t, func() bool { return working(first.str("id")) && working(second.str("id")) }, 10*time.Second, 20*time.Millisecond)
	waiting, err := f.opts.Board.Notes.Get(t.Context(), queued.str("id"))
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, waiting.State, "it waits for the prompt in its directory")
	for _, dir := range []string{images, docs} {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "release"), nil, 0o644))
	}
	for _, id := range []string{first.str("id"), queued.str("id"), second.str("id")} {
		assert.Equal(t, model.NoteDone, f.answered(t, id).State)
	}
	listed := f.call(t, asAdmin, http.MethodGet, "/prompts?folder=infra&dir=images", nil)
	assert.Len(t, listed.list("notes"), 2)
	assert.Equal(t, "again", listed.list("notes")[0].(map[string]any)["body"], "newest first")

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
		{"no parent", api.PromptRequest{Parent: "nothing", Prompt: "x"}, http.StatusNotFound},
	} {
		res := f.call(t, asAdmin, http.MethodPost, "/prompts", c.req)
		assert.Equal(t, c.want, res.Status, c.name+": "+strings.TrimSpace(string(res.Raw)))
	}
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/prompts?folder=nowhere", nil).Status)
}
