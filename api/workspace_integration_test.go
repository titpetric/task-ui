//go:build integration

package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/layout"
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

	// A fresh instance has the default workspace alone.
	empty := f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Equal(t, http.StatusOK, empty.Status)
	assert.Len(t, empty.list("workspaces"), 1)
	assert.Equal(t, model.Default, empty.list("workspaces")[0].(map[string]any)["folder"])

	created := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta", Title: "Malta", Branch: "main"})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	assert.Equal(t, "malta", created.str("workspace", "folder"))
	assert.Equal(t, "malta", created.str("folder"), "the form's data-next fills {folder}")
	assert.Equal(t, repo, created.str("workspace", "home_url"), "the URL clones as the home")
	assert.Equal(t, "main", created.str("workspace", "home_branch"))
	assert.Nil(t, created.get("project"), "no project row: the home is a folder")
	assert.Equal(t, false, created.get("workspace", "shallow"), "Recursive is the default")
	assert.FileExists(t, filepath.Join(f.root, "malta", "home", "Makefile"))
	assert.DirExists(t, filepath.Join(f.root, "malta", "src"))
	assert.NoFileExists(t, filepath.Join(f.root, "malta", "Makefile"))
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/projects", nil).list("projects"), 0)

	// A second workspace from the same repository, under its own name.
	again := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "gozo"})
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))

	list := f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Len(t, list.list("workspaces"), 3)
	assert.Equal(t, "malta", list.list("workspaces")[1].(map[string]any)["folder"], "an added workspace goes last, after the default one")

	got := f.call(t, asAdmin, http.MethodGet, "/workspaces/malta", nil)
	assert.Equal(t, http.StatusOK, got.Status)
	assert.Equal(t, "Malta", got.str("title"))

	updated := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Description: "org tooling", URL: "file:///elsewhere.git"})
	assert.Equal(t, http.StatusOK, updated.Status, string(updated.Raw))
	assert.Equal(t, "The island", updated.str("title"))
	assert.Equal(t, "org tooling", updated.str("description"))
	assert.Equal(t, repo, updated.str("home_url"), "an update keeps the home")
	assert.Equal(t, got.str("id"), updated.str("id"))
	assert.Equal(t, float64(1), updated.get("sort"), "after the default workspace's 0")

	// A sort places the workspace; left out, it stays.
	sort := int64(5)
	sorted := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Sort: &sort})
	assert.Equal(t, http.StatusOK, sorted.Status, string(sorted.Raw))
	assert.Equal(t, float64(5), sorted.get("sort"))
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Equal(t, "gozo", list.list("workspaces")[1].(map[string]any)["folder"])
	kept := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island"})
	assert.Equal(t, float64(5), kept.get("sort"))

	// Recursive off makes the workspace shallow; left out, it stays.
	flat := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Recursive: new(false)})
	assert.Equal(t, true, flat.get("shallow"))
	kept = f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island"})
	assert.Equal(t, true, kept.get("shallow"))
	deep := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Recursive: new(true)})
	assert.Equal(t, false, deep.get("shallow"))

	// A sidebar limit lists the newest projects; left out, it stays; 0
	// lists every one again; below 0 answers 400.
	assert.Equal(t, float64(0), deep.get("sidebar_limit"))
	limit := int64(2)
	limited := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", SidebarLimit: &limit})
	assert.Equal(t, http.StatusOK, limited.Status, string(limited.Raw))
	assert.Equal(t, float64(2), limited.get("sidebar_limit"))
	kept = f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island"})
	assert.Equal(t, float64(2), kept.get("sidebar_limit"))
	assert.Equal(t, float64(2), f.call(t, asAdmin, http.MethodGet, "/workspaces/malta", nil).get("sidebar_limit"))
	limit = -1
	refused := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", SidebarLimit: &limit})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
	assert.Contains(t, refused.str("error"), "sidebar_limit")
	limit = 0
	all := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", SidebarLimit: &limit})
	assert.Equal(t, float64(0), all.get("sidebar_limit"))

	// The hours, trimmed; left out, they stay; empty allows any time;
	// hours the hours package does not read answer 400.
	hours := " mon-fri 9-17\nfri 9-13 "
	gated := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Hours: &hours})
	assert.Equal(t, http.StatusOK, gated.Status, string(gated.Raw))
	assert.Equal(t, "mon-fri 9-17\nfri 9-13", gated.str("hours"))
	kept = f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island"})
	assert.Equal(t, "mon-fri 9-17\nfri 9-13", kept.str("hours"))
	hours = "someday 9-17"
	badHours := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Hours: &hours})
	assert.Equal(t, http.StatusBadRequest, badHours.Status)
	assert.Contains(t, badHours.str("error"), "someday")
	hours = ""
	open := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta", api.WorkspaceRequest{Title: "The island", Hours: &hours})
	assert.Equal(t, "", open.str("hours"))

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
		{"home", api.WorkspaceRequest{URL: repo, Folder: "home"}},
		{"src", api.WorkspaceRequest{Folder: "src"}},
	} {
		res := f.call(t, asAdmin, http.MethodPost, "/workspaces", c.req)
		assert.Equal(t, http.StatusBadRequest, res.Status, c.name)
	}
	internal := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "internal"})
	assert.Contains(t, internal.str("error"), "task-ui's own names")
	assert.NoDirExists(t, filepath.Join(f.root, "internal"))

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)
	duplicate := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"})
	assert.Equal(t, http.StatusConflict, duplicate.Status)

	// stray is a directory under the root already.
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, "stray"), 0o755))
	onDisk := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "stray"})
	assert.Equal(t, http.StatusConflict, onDisk.Status)

	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/projects/gone", api.ProjectRequest{Title: "Gone"}).Status)
	project := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "gone"})
	assert.Equal(t, http.StatusConflict, project.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "src", "gone"))

	missing := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: "file://" + filepath.Join(t.TempDir(), "nothing.git"), Folder: "nothing"})
	assert.Equal(t, http.StatusBadGateway, missing.Status)
	assert.NotEqual(t, "", missing.str("error"))
	assert.NoDirExists(t, filepath.Join(f.root, "src", "nothing"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/nothing", nil).Status)
}

func TestWorkspace_CloneProject(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)

	plain := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{Folder: "malta"})
	assert.Equal(t, http.StatusCreated, plain.Status, string(plain.Raw))
	assert.Equal(t, "", plain.str("workspace", "home_url"), "no URL: an empty home")
	assert.DirExists(t, filepath.Join(f.root, "malta", "home"))
	assert.DirExists(t, filepath.Join(f.root, "malta", "src"))

	res := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "dockerfiles", Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "malta/dockerfiles", res.str("project", "folder"))
	assert.FileExists(t, filepath.Join(f.root, "malta", "src", "dockerfiles", "Makefile"))

	named := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, named.Status, string(named.Raw))
	assert.Equal(t, "malta/website", named.str("project", "folder"), "the ID comes from the URL")

	got := f.call(t, asAdmin, http.MethodGet, "/projects/malta/dockerfiles", nil)
	assert.Equal(t, http.StatusOK, got.Status)

	missing := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "x", Workspace: "gozo"})
	assert.Equal(t, http.StatusBadRequest, missing.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "gozo"))

	for _, id := range []string{"home", "src", "home/x"} {
		taken := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: id, Workspace: "malta"})
		assert.Equal(t, http.StatusBadRequest, taken.Status, id)
	}
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

	// A secret in the workspace, shared or a folder's, stops the delete until someone deletes it.
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:malta", Name: "TOKEN", Content: "shared value"}))
	blocked := f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta", nil)
	assert.Equal(t, http.StatusConflict, blocked.Status, string(blocked.Raw))
	assert.Contains(t, string(blocked.Raw), "holds 1 variables: delete them first")
	assert.DirExists(t, filepath.Join(f.root, "malta"))
	held, err := f.opts.Secrets.Encrypted(ctx, "malta")
	assert.NoError(t, err)
	assert.Len(t, held, 1)
	assert.NoError(t, f.opts.Secrets.Delete(ctx, held[0].ID))

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
	promptCommand(t, f)
	out := promptOut(f, "env.out")

	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)

	res := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: " list the skills ", Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "malta", res.str("dir"), "filed at the workspace; Claude works in its home")
	assert.Equal(t, "list the skills", res.str("body"))
	asked := f.answered(t, res.str("id"))
	run := f.finished(t, asked.RunID)
	assert.Equal(t, model.StateOK, run.State)
	assert.Equal(t, "malta", run.Folder, "the run is the workspace's")
	assert.Equal(t, model.PromptJob, run.Job)
	assert.Equal(t, asAdmin, run.UserName)

	env, err := os.ReadFile(out)
	assert.NoError(t, err)
	dir := filepath.Join(f.root, "malta")
	assert.Equal(t, "TASK_UI_PROMPT=list the skills\nTASK_UI_WORKSPACE=malta\nTASK_UI_WORKSPACE_DIR="+dir+"\npwd="+filepath.Join(dir, "home")+"\n", string(env))

	empty := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: "  "})
	assert.Equal(t, http.StatusBadRequest, empty.Status)
	missing := f.call(t, asAdmin, http.MethodPost, "/workspaces/gozo/prompt", api.PromptRequest{Prompt: "hello"})
	assert.Equal(t, http.StatusNotFound, missing.Status)
	for _, user := range []string{asOps, asViewer, asLimited} {
		refused := f.call(t, user, http.MethodPost, "/workspaces/malta/prompt", api.PromptRequest{Prompt: "hello"})
		assert.Equal(t, http.StatusForbidden, refused.Status, user)
	}
}

// answered waits for a prompt's note to end, and returns it.
func (f *fixture) answered(t *testing.T, id string) *model.Note {
	t.Helper()
	var note *model.Note
	assert.Eventually(t, func() bool {
		var err error
		note, err = f.opts.Board.Notes.Get(context.Background(), id)
		return err == nil && note.Finished()
	}, 10*time.Second, 20*time.Millisecond)
	return note
}

// promptCommand stands a script in for task-ui agent, the command a
// prompt's @claude-prompt runs: it writes the prompt's environment and
// where it runs to promptOut env.out, the run a prompt about a run gets
// to run.out, waits for a file named release in the directory it runs
// in, when that directory holds a file named hold, then answers with the
// prompt in the reply file.
func promptCommand(t *testing.T, f *fixture) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `{ env | grep '^TASK_UI_\(PROMPT\|WORKSPACE\|WORKSPACE_DIR\|PROMPT_DIR\)=' | sort; echo "pwd=$(pwd -P)"; } > "` + promptOut(f, "env.out") + `"; if [ -n "$TASK_UI_PROMPT_RUN" ]; then cp "$TASK_UI_PROMPT_RUN" "` + promptOut(f, "run.out") + `"; fi; while [ -e hold ] && [ ! -e release ]; do sleep 0.05; done; printf 'Answer to: %s' "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}
}

// promptOut is where promptCommand's script writes a file, in a
// directory of the root with nothing to run.
func promptOut(f *fixture, name string) string {
	return filepath.Join(f.root, ".prompt", name)
}

// The default workspace, _, is the root: it lists first, takes a title,
// keeps its place, and stays. No workspace or project takes its name.
func TestWorkspace_Default(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	repo := "file://" + localRepo(t)
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)

	list := f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Len(t, list.list("workspaces"), 2)
	assert.Equal(t, model.Default, list.list("workspaces")[0].(map[string]any)["folder"])

	sort := int64(9)
	updated := f.call(t, asAdmin, http.MethodPut, "/workspaces/_", api.WorkspaceRequest{Title: "Root", Sort: &sort})
	assert.Equal(t, http.StatusOK, updated.Status, string(updated.Raw))
	assert.Equal(t, "Root", updated.str("title"))
	assert.NotEqual(t, float64(9), updated.get("sort"), "the default keeps its place")

	taken := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: model.Default})
	assert.Equal(t, http.StatusBadRequest, taken.Status)
	for _, id := range []string{"_", "_/web"} {
		res := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: id})
		assert.Equal(t, http.StatusBadRequest, res.Status, id)
		assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/projects/"+id, api.ProjectRequest{Title: "Mine"}).Status, id)
	}
	assert.NoDirExists(t, filepath.Join(f.root, model.Default))

	// _ clones into src at the root, as a workspace does into its own.
	root := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "site", Workspace: model.Default})
	assert.Equal(t, http.StatusCreated, root.Status, string(root.Raw))
	assert.Equal(t, "site", root.str("project", "folder"))
	assert.FileExists(t, filepath.Join(f.root, "src", "site", "Makefile"))
	assert.DirExists(t, filepath.Join(f.root, "home"), "the root has an empty home")

	// Deleting _ deletes it like any other workspace: 409 while a run
	// of the root is active and while it holds a secret; then its
	// project rows and checkouts go, a checkout without a row too, its
	// home with its checkout, and its row; the root directory and the
	// other workspaces stay.
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/projects/plain", api.ProjectRequest{Title: "Plain"}).Status)
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, "src", "loose", ".git"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(f.root, "home", ".profile"), []byte("export A=1\n"), 0o644))
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/workspaces/_/home", api.WorkspaceHomeRequest{URL: repo}).Status)
	slow := f.start(t, asAdmin, "infra", "slow")
	busy := f.call(t, asAdmin, http.MethodDelete, "/workspaces/_", nil)
	assert.Equal(t, http.StatusConflict, busy.Status, string(busy.Raw))
	assert.Contains(t, busy.str("error"), "infra/slow is running in internal")
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodPost, "/runs/"+slow+"/kill", nil).Status)
	f.finished(t, slow)
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:internal", Name: "TOKEN", Content: "shared value"}))
	blocked := f.call(t, asAdmin, http.MethodDelete, "/workspaces/_", nil)
	assert.Equal(t, http.StatusConflict, blocked.Status, string(blocked.Raw))
	assert.Contains(t, blocked.str("error"), "holds 1 variables: delete them first")
	held, err := f.opts.Secrets.Encrypted(ctx, model.Default)
	assert.NoError(t, err)
	assert.Len(t, held, 1)
	assert.NoError(t, f.opts.Secrets.Delete(ctx, held[0].ID))

	deleted := f.call(t, asAdmin, http.MethodDelete, "/workspaces/_", nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status, string(deleted.Raw))
	assert.DirExists(t, f.root)
	assert.DirExists(t, filepath.Join(f.root, "malta", "home"), "a named workspace stays")
	assert.NoDirExists(t, filepath.Join(f.root, "home"))
	assert.NoDirExists(t, filepath.Join(f.root, "src"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/_", nil).Status)
	projects, err := f.opts.Projects.List(ctx)
	assert.NoError(t, err)
	assert.Empty(t, projects, "the root's project rows go")
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces", nil)
	assert.Len(t, list.list("workspaces"), 1)
	assert.Equal(t, "malta", list.list("workspaces")[0].(map[string]any)["folder"])

	// A restart brings nothing back: the row stays gone and the root
	// gets no home or src.
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	rows, err := f.opts.Workspaces.List(ctx)
	assert.NoError(t, err)
	for _, row := range rows {
		assert.NoError(t, layout.Ensure(f.root, row.Folder))
	}
	assert.Len(t, rows, 1)
	assert.NoDirExists(t, filepath.Join(f.root, "home"))
	assert.NoDirExists(t, filepath.Join(f.root, "src"))

	// A clone needs a workspace now.
	none := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "site"})
	assert.Equal(t, http.StatusBadRequest, none.Status, string(none.Raw))
	assert.Equal(t, "pick a workspace: the instance has no default workspace", none.str("error"))
	assert.NoDirExists(t, filepath.Join(f.root, "src"))
	inMalta := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "site", Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, inMalta.Status, string(inMalta.Raw))
	assert.FileExists(t, filepath.Join(f.root, "malta", "src", "site", "Makefile"))
}

// A project cloned into the root may not take a workspace's name, or
// one a directory holds, and git does not run: a repository that is not
// there would answer 502.
func TestWorkspace_RootCloneTaken(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)
	nothing := "file://" + filepath.Join(t.TempDir(), "nothing.git")
	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"}).Status)

	for _, id := range []string{"malta", "malta/x", "infra"} {
		res := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: nothing, ID: id})
		assert.Equal(t, http.StatusConflict, res.Status, id)
		assert.Equal(t, "", res.str("output"), id)
	}
	assert.NoDirExists(t, filepath.Join(f.root, "malta", "src", "x"))
}

// Replace clones a repository over a workspace's home, Remove
// empties it, the projects beside it stay, and both refuse while a run
// of the workspace is active: the home is the HOME of every run in it.
func TestWorkspace_Home(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	repo := "file://" + localRepo(t)
	assert.Equal(t, http.StatusCreated, f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{Folder: "malta"}).Status)
	a := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "a", Workspace: "malta"})
	assert.Equal(t, http.StatusCreated, a.Status, string(a.Raw))
	home := filepath.Join(f.root, "malta", "home")
	url := func() string {
		return f.call(t, asAdmin, http.MethodGet, "/workspaces/malta", nil).str("home_url")
	}

	assert.Equal(t, "", url())
	assert.NoFileExists(t, filepath.Join(home, "Makefile"))
	replaced := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/home", api.WorkspaceHomeRequest{URL: repo, Branch: "main"})
	assert.Equal(t, http.StatusOK, replaced.Status, string(replaced.Raw))
	assert.Equal(t, repo, replaced.str("workspace", "home_url"))
	assert.Equal(t, "main", replaced.str("workspace", "home_branch"))
	assert.Equal(t, repo, url())
	assert.FileExists(t, filepath.Join(home, "Makefile"))
	assert.FileExists(t, filepath.Join(f.root, "malta", "src", "a", "Makefile"), "the project stays")

	for _, c := range []struct{ name, url string }{{"no url", ""}, {"not a url", "not a url"}} {
		refused := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/home", api.WorkspaceHomeRequest{URL: c.url})
		assert.Equal(t, http.StatusBadRequest, refused.Status, c.name)
	}
	assert.Equal(t, repo, url(), "a refusal keeps the home")
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPut, "/workspaces/gozo/home", api.WorkspaceHomeRequest{URL: repo}).Status)

	// A run in the workspace holds the home.
	f.addFolder("malta/a", shellFolder(t, f.root, "malta/a", map[string]string{"slow": "sleep 30"}))
	slow := f.start(t, asAdmin, "malta/a", "slow")
	busy := f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/home", nil)
	assert.Equal(t, http.StatusConflict, busy.Status, string(busy.Raw))
	assert.Contains(t, busy.str("error"), "malta/a/slow")
	assert.FileExists(t, filepath.Join(home, "Makefile"))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodPost, "/runs/"+slow+"/kill", nil).Status)
	f.finished(t, slow)

	removed := f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/home", nil)
	assert.Equal(t, http.StatusOK, removed.Status, string(removed.Raw))
	assert.Equal(t, "", url(), "Remove")
	assert.DirExists(t, home, "the home stays, empty")
	assert.NoFileExists(t, filepath.Join(home, "Makefile"))
	entries, err := os.ReadDir(home)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)

	// A clone that fails leaves an empty home and no URL.
	failed := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/home", api.WorkspaceHomeRequest{URL: "file://" + filepath.Join(t.TempDir(), "nothing.git")})
	assert.Equal(t, http.StatusBadGateway, failed.Status, string(failed.Raw))
	assert.DirExists(t, home)
	assert.Equal(t, "", url())
	ws, err := f.opts.Workspaces.GetByFolder(ctx, "malta")
	assert.NoError(t, err)
	assert.Equal(t, "", ws.HomeURL)

	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.Equal(t, http.StatusForbidden, f.call(t, user, http.MethodPut, "/workspaces/malta/home", api.WorkspaceHomeRequest{URL: repo}).Status, user)
		assert.Equal(t, http.StatusForbidden, f.call(t, user, http.MethodDelete, "/workspaces/malta/home", nil).Status, user)
	}
}
