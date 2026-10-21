package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestWorkspace_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/workspaces"},
		{http.MethodPost, "/workspaces"},
		{http.MethodGet, "/workspaces/malta"},
		{http.MethodPut, "/workspaces/malta"},
		{http.MethodDelete, "/workspaces/malta"},
		{http.MethodPost, "/workspaces/malta/prompt"},
	} {
		for _, user := range []string{asOps, asViewer, asLimited} {
			res := f.call(t, user, c.method, c.path, map[string]any{})
			assert.Equal(t, http.StatusForbidden, res.Status, user+" "+c.method+" "+c.path)
		}
	}
}

func TestWorkspace_CRUD(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)

	empty := f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Equal(t, http.StatusOK, empty.Status)
	assert.Contains(t, string(empty.Raw), `"workspaces":[]`)

	created := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta", Title: "Malta", Branch: "main"})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	assert.Equal(t, "malta", created.str("workspace", "folder"))
	assert.Equal(t, repo, created.str("workspace", "source_url"))
	assert.FileExists(t, filepath.Join(f.root, "malta", "Makefile"))

	// A second workspace from the same repository, under its own name.
	again := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "gozo"})
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))

	list := f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Len(t, list.list("workspaces"), 2)
	assert.Equal(t, "malta", list.list("workspaces")[0].(map[string]any)["folder"], "an added workspace goes last")

	got := f.call(t, asAdmin, http.MethodGet, "/workspaces/malta", nil)
	assert.Equal(t, http.StatusOK, got.Status)
	assert.Equal(t, "Malta", got.str("title"))

	updated := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Description: "org tooling", URL: "file:///elsewhere.git"})
	assert.Equal(t, http.StatusOK, updated.Status, string(updated.Raw))
	assert.Equal(t, "The island", updated.str("title"))
	assert.Equal(t, "org tooling", updated.str("description"))
	assert.Equal(t, repo, updated.str("source_url"), "an update keeps the repository")
	assert.Equal(t, got.str("id"), updated.str("id"))
	assert.Equal(t, float64(0), updated.get("sort"))

	// A sort places the workspace; left out, it stays.
	sort := int64(5)
	sorted := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Sort: &sort})
	assert.Equal(t, http.StatusOK, sorted.Status, string(sorted.Raw))
	assert.Equal(t, float64(5), sorted.get("sort"))
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Equal(t, "gozo", list.list("workspaces")[0].(map[string]any)["folder"])
	kept := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island"})
	assert.Equal(t, float64(5), kept.get("sort"))

	deleted := f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta", nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "malta"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/malta", nil).Status)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		res := f.call(t, asAdmin, method, "/workspaces/missing", api.WorkspaceRequest{})
		assert.Equal(t, http.StatusNotFound, res.Status, method)
	}
}

func TestWorkspace_CreateRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)

	for _, c := range []struct {
		name string
		req  api.WorkspaceRequest
	}{
		{"no url", api.WorkspaceRequest{URL: "not a url", Folder: "malta"}},
		{"no folder", api.WorkspaceRequest{URL: repo}},
		{"nested folder", api.WorkspaceRequest{URL: repo, Folder: "malta/x"}},
		{"unsafe folder", api.WorkspaceRequest{URL: repo, Folder: ".."}},
		{"internal", api.WorkspaceRequest{URL: repo, Folder: "internal"}},
	} {
		res := f.call(t, asAdmin, http.MethodPost, "/workspaces", c.req)
		assert.Equal(t, http.StatusBadRequest, res.Status, c.name)
	}
	internal := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "internal"})
	assert.Contains(t, internal.str("error"), "task-ui's own namespace")
	assert.NoDirExists(t, filepath.Join(f.root, "internal"))

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)
	duplicate := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"})
	assert.Equal(t, http.StatusConflict, duplicate.Status)

	// infra is a directory under the root already.
	onDisk := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "infra"})
	assert.Equal(t, http.StatusConflict, onDisk.Status)

	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/projects/gone", api.ProjectRequest{Title: "Gone"}).Status)
	project := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "gone"})
	assert.Equal(t, http.StatusConflict, project.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "gone"))

	missing := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: "file://" + filepath.Join(t.TempDir(), "nothing.git"), Folder: "nothing"})
	assert.Equal(t, http.StatusBadGateway, missing.Status)
	assert.NotEqual(t, "", missing.str("error"))
	assert.NoDirExists(t, filepath.Join(f.root, "nothing"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/nothing", nil).Status)
}

func TestWorkspace_CloneProject(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)

	res := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "dockerfiles", Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "malta/dockerfiles", res.str("project", "folder"))
	assert.FileExists(t, filepath.Join(f.root, "malta", "dockerfiles", "Makefile"))

	named := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, named.Status, string(named.Raw))
	assert.Equal(t, "malta/website", named.str("project", "folder"), "the ID comes from the URL")

	got := f.call(t, asAdmin, http.MethodGet, "/projects/malta/dockerfiles", nil)
	assert.Equal(t, http.StatusOK, got.Status)

	missing := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "x", Workspace: "gozo"})
	assert.Equal(t, http.StatusBadRequest, missing.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "gozo"))

	internal := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "x", Workspace: "internal"})
	assert.Equal(t, http.StatusBadRequest, internal.Status)
	assert.Contains(t, internal.str("error"), "task-ui's own namespace")
	assert.NoDirExists(t, filepath.Join(f.root, "internal"))
}

func TestWorkspace_DeleteNested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)
	ctx := context.Background()

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)
	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "dockerfiles", Workspace: "malta"}).Status)
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/projects/malta/dockerfiles/sub", api.ProjectRequest{Title: "Sub"}).Status)
	// A project whose name only starts like the workspace's stays.
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/projects/maltase", api.ProjectRequest{Title: "Other"}).Status)

	res := f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta", nil)
	assert.Equal(t, http.StatusNoContent, res.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "malta"))

	projects, err := f.opts.Projects.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "maltase", projects[0].Folder)

	// A workspace whose checkout went keeps no rows either.
	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "gozo"}).Status)
	assert.NoError(t, os.RemoveAll(filepath.Join(f.root, "gozo")))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/gozo", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/gozo", nil).Status)
}

func TestWorkspace_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)
	// The internal job stands in for task-ui agent prompt: it writes its
	// environment to a file.
	internal := shellFolder(t, f.root, model.Internal, map[string]string{
		"claude-prompt": `env | grep '^TASK_UI_\(PROMPT\|WORKSPACE\)' | sort > "$PWD/env.out"`,
	})
	internal.Job("claude-prompt").Allow = []string{model.RoleAdmin}
	f.opts.Catalogue.(catalogue)[model.Internal] = internal

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)

	res := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: " list the skills ", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, model.Internal, res.str("folder"))
	assert.Equal(t, "claude-prompt", res.str("job"))
	run := f.finished(t, res.str("id"))
	assert.Equal(t, model.StateOK, run.State)
	assert.Equal(t, asAdmin, run.UserName)

	env, err := os.ReadFile(filepath.Join(internal.Dir, "env.out"))
	assert.NoError(t, err)
	dir := filepath.Join(f.root, "malta")
	assert.Equal(t, "TASK_UI_PROMPT=list the skills\nTASK_UI_WORKSPACE=malta\nTASK_UI_WORKSPACE_DIR="+dir+"\n", string(env))

	empty := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: "  "})
	assert.Equal(t, http.StatusBadRequest, empty.Status)
	missing := f.call(t, asAdmin, http.MethodPost, "/workspaces/gozo/prompt", api.PromptRequest{Prompt: "hello"})
	assert.Equal(t, http.StatusNotFound, missing.Status)
	for _, user := range []string{asOps, asViewer, asLimited} {
		refused := f.call(t, user, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: "hello"})
		assert.Equal(t, http.StatusForbidden, refused.Status, user)
	}
}
