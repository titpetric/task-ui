//go:build integration

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
// directory inside one: a note, which a run of @claude-prompt in that
// folder and directory answers, with TASK_UI_PROMPT_DIR naming it, as a
// reply. Admins only.
func TestPrompts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptCommand(t, f)
	out := promptOut(f, "env.out")
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
	assert.Equal(t, "infra", run.Folder)
	assert.Equal(t, model.PromptJob, run.Job)
	assert.Equal(t, "", run.Dir)
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

	// Prompts go side by side: another in the same directory runs at
	// once, as one in another directory does.
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
	beside := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "images/", Prompt: "again"})
	assert.Equal(t, http.StatusCreated, beside.Status, string(beside.Raw))
	second := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Dir: "docs", Prompt: "read"})
	assert.Equal(t, http.StatusCreated, second.Status, string(second.Raw))
	working := func(id string) bool {
		note, err := f.opts.Board.Notes.Get(t.Context(), id)
		return err == nil && note.State == model.NoteWorking
	}
	assert.Eventually(t, func() bool {
		return working(first.str("id")) && working(beside.str("id")) && working(second.str("id"))
	}, 10*time.Second, 20*time.Millisecond)
	for _, dir := range []string{images, docs} {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "release"), nil, 0o644))
	}
	for _, id := range []string{first.str("id"), beside.str("id"), second.str("id")} {
		assert.Equal(t, model.NoteDone, f.answered(t, id).State)
	}
	imagesRun := f.finished(t, f.answered(t, first.str("id")).RunID)
	assert.Equal(t, "infra/images/"+model.PromptJob, imagesRun.Key(), "the run's job is folder/dir/@claude-prompt")
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

// A runner with prompt/edit asks too, reads the run of the prompt and
// ends it. Others who see the folder see a prompt's run with
// prompt/view only: in /api/runs, its page and the runs topic. No
// prompt starts from POST /api/runs.
func TestPrompts_ByPermission(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptCommand(t, f)
	ctx := t.Context()
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asOps, "prompt/edit", "allow"))

	res := f.call(t, asOps, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "what runs here"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	asked := f.answered(t, res.str("id"))
	assert.Equal(t, model.NoteDone, asked.State)
	assert.Equal(t, asOps, f.finished(t, asked.RunID).UserName)

	assert.Equal(t, http.StatusOK, f.call(t, asOps, http.MethodGet, "/runs/"+asked.RunID, nil).Status)
	assert.Equal(t, http.StatusOK, f.call(t, asOps, http.MethodGet, "/runs/"+asked.RunID+"/frame.txt", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asViewer, http.MethodGet, "/runs/"+asked.RunID, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/runs/"+asked.RunID, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodPost, "/runs/"+asked.RunID+"/stop", nil).Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asOps, http.MethodPost, "/runs/"+asked.RunID+"/stop", nil).Status)
	listed := func(user string) bool {
		for _, run := range f.call(t, user, http.MethodGet, "/runs?folder=infra", nil).list("runs") {
			if run.(map[string]any)["id"] == asked.RunID {
				return true
			}
		}
		return false
	}
	assert.True(t, listed(asOps), "the asker's own")
	assert.False(t, listed(asLimited), "without prompt/view")
	assert.True(t, listed(asAdmin))
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asLimited, "prompt/view", "allow"))
	assert.True(t, listed(asLimited), "with prompt/view")
	assert.Equal(t, http.StatusOK, f.call(t, asLimited, http.MethodGet, "/runs/"+asked.RunID, nil).Status)

	refused := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: model.PromptJob})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
}

// POST /api/prompts with run asks about a run, in its folder and
// directory: Claude gets the run, the job's command and the end of its
// output, cut to 2000 lines, and the thread lists by the run, apart
// from its directory's prompts.
func TestPrompts_AboutRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptCommand(t, f)
	f.opts.Catalogue.(catalogue)["logs"] = shellFolder(t, f.root, "logs", map[string]string{"long": "seq 1 2600; exit 4"})
	about := f.finished(t, f.start(t, asAdmin, "logs", "long"))
	assert.Equal(t, model.StateFailed, about.State)

	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodPost, "/prompts", api.PromptRequest{Run: about.ID, Prompt: "why"})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
		assert.Equal(t, http.StatusForbidden, f.call(t, user, http.MethodGet, "/prompts?run="+about.ID, nil).Status, user)
	}

	res := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: about.ID, Prompt: "why did it end so"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "logs", res.str("dir"))
	assert.Equal(t, about.ID, res.str("about_run_id"))
	asked := f.answered(t, res.str("id"))
	assert.Equal(t, model.NoteDone, asked.State)
	prompt := f.finished(t, asked.RunID)
	assert.Equal(t, "logs/"+model.PromptJob, prompt.Key())
	refused := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: prompt.ID, Prompt: "why"})
	assert.Equal(t, http.StatusNotFound, refused.Status, "no prompt about a prompt's run")

	text, err := os.ReadFile(promptOut(f, "run.out"))
	assert.NoError(t, err)
	got := string(text)
	assert.Contains(t, got, "Run "+about.ID+" of the job long in logs, by admin.\nState: failed, exit code 4.\n")
	assert.Contains(t, got, "The job's command: sh -c seq 1 2600; exit 4\n")
	assert.Contains(t, got, "cut to its last 2000 lines or 256 KiB:\n\n```\n601\n602\n")
	assert.Contains(t, got, "\n2600\n```\n")
	assert.NotContains(t, got, "\n600\n")

	listed := f.call(t, asAdmin, http.MethodGet, "/prompts?run="+about.ID, nil)
	assert.Equal(t, http.StatusOK, listed.Status, string(listed.Raw))
	assert.Len(t, listed.list("notes"), 1)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/prompts?folder=logs", nil).list("notes"), 0, "the directory lists its own prompts")

	// A reply asks again about the same run.
	assert.NoError(t, os.Remove(promptOut(f, "run.out")))
	again := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Parent: asked.ID, Prompt: "and then"})
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))
	assert.Equal(t, about.ID, again.str("about_run_id"))
	assert.Eventually(t, func() bool {
		replies := f.call(t, asAdmin, http.MethodGet, "/prompts?run="+about.ID, nil).list("notes")[0].(map[string]any)["replies"].([]any)
		return len(replies) == 3
	}, 10*time.Second, 20*time.Millisecond)
	assert.FileExists(t, promptOut(f, "run.out"))

	// A run the user does not see, or one of the internal folder, has no
	// prompt.
	for _, rule := range []string{"prompt/view", "prompt/edit"} {
		assert.NoError(t, f.opts.Users.Policies().Put(t.Context(), "user", asLimited, rule, "allow"))
	}
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodPost, "/prompts", api.PromptRequest{Run: about.ID, Prompt: "x"}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/prompts?run="+about.ID, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: asked.RunID, Prompt: "x"}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: "nothing", Prompt: "x"}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: about.ID, Prompt: " "}).Status)
}

// A prompt of the area workflows runs in the root, or the folder it
// names, with the builder's instruction, and lists by the area, apart
// from the root's prompts; an area there is none of answers 400.
func TestPrompts_Area(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `printf '%s' "$TASK_UI_PROMPT_INSTRUCTIONS" > "$TASK_UI_PROMPT_REPLY"`}

	res := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Area: "workflows", Prompt: "build a weather workflow"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, model.Default, res.str("dir"))
	assert.Equal(t, model.AreaWorkflows, res.str("area"))
	asked := f.answered(t, res.str("id"))
	assert.Equal(t, model.NoteDone, asked.State)

	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?area=workflows", nil).list("notes")
	assert.Len(t, threads, 1)
	answer := threads[0].(map[string]any)["replies"].([]any)[0].(map[string]any)["body"].(string)
	assert.Contains(t, answer, "create_workflow or update_workflow")
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/prompts?folder=_", nil).list("notes"), 0)

	inFolder := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Area: "workflows", Folder: "infra", Prompt: "build one here"})
	assert.Equal(t, http.StatusCreated, inFolder.Status, string(inFolder.Raw))
	assert.Equal(t, "infra", inFolder.str("dir"))
	f.answered(t, inFolder.str("id"))

	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Area: "elsewhere", Prompt: "x"}).Status)
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/prompts", api.PromptRequest{Area: "workflows", Prompt: "x"}).Status)
}
