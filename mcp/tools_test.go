package mcp_test

import (
	"context"
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

// The tools list a job, run it, wait for it and show it; a viewer may not run it.
func TestTools_Jobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "admin", "list_jobs", nil)
	assert.False(t, failed)
	assert.Contains(t, text, "site\n  build - Build the site")

	text, failed = f.tool(t, "guest", "run_job", map[string]any{"folder": "site", "job": "build"})
	assert.True(t, failed)
	assert.Contains(t, text, "not run it (403)")

	text, failed = f.tool(t, "admin", "run_job", map[string]any{"folder": "site", "job": "build", "wait": 20})
	assert.False(t, failed, text)
	assert.Contains(t, text, "site/build by admin, ok (exit 0)")
	assert.Contains(t, text, "built")
	id := strings.Fields(strings.TrimPrefix(text, "run "))[0]
	id = strings.TrimSuffix(id, ":")

	text, _ = f.tool(t, "guest", "list_jobs", map[string]any{"folder": "site"})
	assert.Contains(t, text, "[view only] | last: ok (exit 0)")

	text, _ = f.tool(t, "guest", "list_runs", map[string]any{"folder": "site", "limit": 5})
	assert.Contains(t, text, id+" site/build by admin: ok")

	text, failed = f.tool(t, "guest", "get_run", map[string]any{"id": id, "lines": 3})
	assert.False(t, failed, text)
	assert.Contains(t, text, "run "+id)

	_, failed = f.tool(t, "admin", "get_run", map[string]any{"id": "nothing"})
	assert.True(t, failed)

	// An interactive shell is an admin's alone.
	for _, user := range []string{"ops", "guest"} {
		text, failed = f.tool(t, user, "run_job", map[string]any{"folder": "site", "job": "@shell"})
		assert.True(t, failed, user)
		assert.Contains(t, text, "(403)", user)
	}
}

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

// The tools list a workflow, run it by name, wait for it and list its runs.
func TestTools_Workflows(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["admin"], "/api/workflows", api.WorkflowRequest{Name: "Ship", Steps: []api.StepRequest{{Folder: "site", Job: "build"}}})
	assert.Equal(t, http.StatusCreated, status, string(raw))

	text, _ := f.tool(t, "admin", "list_workflows", nil)
	assert.Contains(t, text, "Ship (id ")
	assert.Contains(t, text, "): site/build")

	text, failed := f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "ship", "wait": 20})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Ship by admin, ok")
	assert.Contains(t, text, "1. site/build: ok (run ")

	text, _ = f.tool(t, "admin", "list_workflow_runs", map[string]any{"workflow": "Ship"})
	assert.Contains(t, text, "Ship by admin, ok")

	_, failed = f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "nothing"})
	assert.True(t, failed)
	_, failed = f.tool(t, "guest", "run_workflow", map[string]any{"workflow": "Ship"})
	assert.True(t, failed)
}

// A note naming the agent is the agent's to list; it replies and
// closes it with note_comment.
func TestTools_Notes(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@admin tidy the README"})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@guest have a look"})
	// A prompt from a folder page is a note too, which the board leaves out.
	assert.NoError(t, f.notes.Create(context.Background(), &model.Note{UserName: "admin", Body: "@admin a prompt", State: model.NoteOpen, WorkspaceID: model.Default, Dir: "site"}))

	text, failed := f.tool(t, "admin", "list_notes", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "by ops, open,")
	assert.Contains(t, text, "@admin tidy the README")
	assert.NotContains(t, text, "@guest")
	assert.NotContains(t, text, "a prompt")
	id := strings.Fields(strings.TrimPrefix(text, "note "))[0]

	text, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id, "body": "Tidied in abc123.", "state": "done"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "it is done")

	text, _ = f.tool(t, "admin", "list_notes", nil)
	assert.Equal(t, "No notes.", text)
	text, _ = f.tool(t, "admin", "list_notes", map[string]any{"state": "*", "mention": "*"})
	assert.Contains(t, text, "reply by admin")
	assert.Contains(t, text, "@guest have a look")
	assert.NotContains(t, text, "a prompt")

	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id})
	assert.True(t, failed)
	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": "nothing", "body": "x"})
	assert.True(t, failed)
}

// A note naming a user other than the agent is that user's to list by
// default and to report on; someone else may not set its state.
func TestTools_NotesToUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@consuela tidy the README"})
	assert.Equal(t, http.StatusCreated, status, string(raw))

	text, failed := f.tool(t, "consuela", "list_notes", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "by ops, open,")
	assert.Contains(t, text, "@consuela tidy the README")
	id := strings.Fields(strings.TrimPrefix(text, "note "))[0]

	text, failed = f.tool(t, "guest", "note_comment", map[string]any{"id": id, "state": "done"})
	assert.True(t, failed, text)
	text, failed = f.tool(t, "consuela", "note_comment", map[string]any{"id": id, "state": "working"})
	assert.False(t, failed, text)
	status, raw = f.post(t, f.tokens["ops"], "/api/notes/"+id+"/comments", map[string]string{"state": "done"})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	text, failed = f.tool(t, "consuela", "note_comment", map[string]any{"id": id, "body": "Tidied in abc123.", "state": "done"})
	assert.False(t, failed, text)
	text, _ = f.tool(t, "consuela", "list_notes", nil)
	assert.Equal(t, "No notes.", text)
}
