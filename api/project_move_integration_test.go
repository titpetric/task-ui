//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// inWorkspace names a workspace for a project request.
func inWorkspace(ws string) api.ProjectRequest {
	return api.ProjectRequest{Workspace: &ws}
}

// withWorkspace makes a workspace row and its checkout under the root.
func (f *fixture) withWorkspace(t *testing.T, ws string) {
	t.Helper()
	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: ws}))
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, ws), 0o755))
}

func TestProject_Move(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.withWorkspace(t, "acme")
	assert.NoError(t, f.opts.Schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan", Expression: "@hourly"}))
	run := f.start(t, asAdmin, "infra", "plan")
	f.finished(t, run)
	before := f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra"})
	assert.Equal(t, http.StatusOK, before.Status)

	res := f.call(t, asAdmin, http.MethodPut, "/projects/infra", inWorkspace("acme"))
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var moved api.ProjectMoved
	assert.NoError(t, json.Unmarshal(res.Raw, &moved))
	assert.Equal(t, "acme/infra", moved.Folder)
	assert.Equal(t, before.str("id"), moved.ID, "the row moves, it is not made anew")
	assert.Len(t, moved.Left, 1)
	assert.Equal(t, "schedule infra · plan", moved.Left[0].String())
	assert.DirExists(t, filepath.Join(f.root, "acme", "infra", "edge"), "the checkout moves with the folders inside it")
	assert.NoDirExists(t, filepath.Join(f.root, "infra"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/projects/infra", nil).Status)

	limited, err := f.opts.Users.Get(ctx, f.users[asLimited].ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme/infra"}, limited.FolderList(), "the grant follows the project")
	old, err := f.opts.Runs.Get(ctx, run)
	assert.NoError(t, err)
	assert.Equal(t, "infra", old.Folder, "the runs stay under the old ID")

	same := f.call(t, asAdmin, http.MethodPut, "/projects/acme/infra", inWorkspace("acme"))
	assert.Equal(t, http.StatusOK, same.Status, string(same.Raw))
	assert.Nil(t, same.get("left"), "naming the workspace it is in saves without a move")

	back := f.call(t, asAdmin, http.MethodPut, "/projects/acme/infra", inWorkspace(model.Default))
	assert.Equal(t, http.StatusOK, back.Status, string(back.Raw))
	assert.Equal(t, "infra", back.str("folder"))
	assert.Len(t, back.list("left"), 0, "the schedule names the ID it has again")
	assert.DirExists(t, filepath.Join(f.root, "infra", "edge"))
	limited, err = f.opts.Users.Get(ctx, f.users[asLimited].ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"infra"}, limited.FolderList())
}

func TestProject_MoveConflicts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.withWorkspace(t, "acme")
	f.withWorkspace(t, "taken")
	assert.NoError(t, os.MkdirAll(filepath.Join(f.root, "taken", "infra"), 0o755))

	for _, c := range []struct {
		name, path, ws string
		status         int
	}{
		{"a workspace", "/projects/acme", model.Default, http.StatusConflict},
		{"no such workspace", "/projects/infra", "nowhere", http.StatusConflict},
		{"the folder is taken", "/projects/infra", "taken", http.StatusConflict},
		{"a workspace holds the name at the root", "/projects/taken/infra", model.Default, http.StatusConflict},
		{"a folder inside a project", "/projects/infra/edge", "acme", http.StatusBadRequest},
		{"into internal", "/projects/infra", model.Internal, http.StatusBadRequest},
	} {
		res := f.call(t, asAdmin, http.MethodPut, c.path, inWorkspace(c.ws))
		assert.Equal(t, c.status, res.Status, c.name+": "+string(res.Raw))
	}
	assert.DirExists(t, filepath.Join(f.root, "infra"))

	held := &model.Project{Folder: "acme/infra"}
	assert.NoError(t, f.opts.Projects.Create(ctx, held))
	row := f.call(t, asAdmin, http.MethodPut, "/projects/infra", inWorkspace("acme"))
	assert.Equal(t, http.StatusConflict, row.Status, "a project row holds the ID")
	assert.NoError(t, f.opts.Projects.Delete(ctx, held.ID))

	slow := f.start(t, asAdmin, "infra", "slow")
	busy := f.call(t, asAdmin, http.MethodPut, "/projects/infra", inWorkspace("acme"))
	assert.Equal(t, http.StatusConflict, busy.Status, string(busy.Raw))
	assert.Contains(t, busy.str("error"), "run")
	assert.DirExists(t, filepath.Join(f.root, "infra"))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodPost, "/runs/"+slow+"/kill", nil).Status)
	f.finished(t, slow)

	free := f.call(t, asAdmin, http.MethodPut, "/projects/infra", inWorkspace("acme"))
	assert.Equal(t, http.StatusOK, free.Status, "once the run ends it moves: "+string(free.Raw))
}
