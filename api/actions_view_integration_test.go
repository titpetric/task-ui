//go:build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills/skillstest"
)

// skillFixture makes infra a checkout with a directory with a compose file,
// and the stand-in atkins, and returns that directory.
func skillFixture(t *testing.T, f *fixture) string {
	t.Helper()
	probe, _ := skillstest.Prober(t)
	f.opts.Catalogue.Find("infra").Probe = probe
	app := filepath.Join(f.root, "infra", "app")
	out, err := exec.Command("git", "-c", "init.defaultBranch=main", "init", "-q", filepath.Join(f.root, "infra")).CombinedOutput()
	assert.NoError(t, err, string(out))
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "compose.yml"), []byte("services:\n  a:\n    build: .\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "atkins.yml"), []byte("jobs:\n  default:\n    steps:\n      - run: echo hi\n"), 0o644))
	return app
}

// GET /folders/{id}/actions?dir= answers the actions bar of a place: the
// git line, the skills atkins applies there with each job's can_run, and
// the place's own jobs. Anyone who sees the folder reads it; nobody
// reads a directory the folder's Files would not open.
func TestFolders_Actions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	skillFixture(t, f)

	res := f.call(t, asViewer, http.MethodGet, "/folders/infra/actions?dir=app", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.ActionsView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "infra", view.Folder)
	assert.Equal(t, "app", view.Dir)
	assert.NotNil(t, view.Git)
	assert.Equal(t, "infra", view.Git.Checkout)
	assert.Equal(t, "main", view.Git.Branch)
	assert.Len(t, view.Skills, 2)
	assert.Equal(t, "git", view.Skills[0].ID)
	assert.Equal(t, "", view.Skills[0].Dir)
	assert.Equal(t, "compose", view.Skills[1].ID)
	assert.Equal(t, "app", view.Skills[1].Dir)
	assert.Equal(t, "@compose:build", view.Skills[1].Jobs[0].ID)
	assert.False(t, view.Skills[1].Jobs[0].CanRun, "a viewer runs nothing")
	assert.Len(t, view.Jobs, 1)
	assert.Equal(t, "default", view.Jobs[0].ID)
	assert.Equal(t, "app", view.Jobs[0].Dir)

	ops := f.call(t, asOps, http.MethodGet, "/folders/infra/actions?dir=app", nil)
	assert.NoError(t, json.Unmarshal(ops.Raw, &view))
	assert.True(t, view.Skills[1].Jobs[0].CanRun)

	top := f.call(t, asOps, http.MethodGet, "/folders/infra/actions", nil)
	assert.Equal(t, http.StatusOK, top.Status)
	assert.Equal(t, "", top.str("dir"))
	assert.Len(t, top.list("skills"), 1)
	assert.Len(t, top.list("jobs"), 6)

	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/folders/private/actions", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/folders/infra/actions?dir=..", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/folders/infra/actions?dir=.git", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/folders/infra", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/folders/nowhere/actions", nil).Status)
}

// A skill job runs atkins skill:job in the directory of the file its
// skill matched, under the folder's rules: git:status from a directory
// runs in the checkout, compose:up in the directory, and the row keeps
// that directory, so Run again finds the job there.
func TestRun_Skill(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	app := skillFixture(t, f)
	checkout, err := filepath.EvalSymlinks(filepath.Join(f.root, "infra"))
	assert.NoError(t, err)

	res := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Dir: "app", Job: "@git:status", Cols: 250, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	run := f.finished(t, res.str("id"))
	assert.Equal(t, model.StateOK, run.State)
	assert.Equal(t, "", run.Dir)
	assert.Equal(t, "@git:status", run.Job)
	text := f.call(t, asViewer, http.MethodGet, "/runs/"+run.ID+"/frame.txt", nil)
	assert.Contains(t, string(text.Raw), "ran git:status in "+checkout)

	up := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Dir: "app", Job: "@compose:up", Cols: 250, Rows: 24})
	assert.Equal(t, http.StatusCreated, up.Status, string(up.Raw))
	row := f.finished(t, up.str("id"))
	assert.Equal(t, "app", row.Dir)
	assert.Equal(t, "infra/app/@compose:up", row.Key())
	real, err := filepath.EvalSymlinks(app)
	assert.NoError(t, err)
	text = f.call(t, asViewer, http.MethodGet, "/runs/"+row.ID+"/frame.txt", nil)
	assert.Contains(t, string(text.Raw), "ran compose:up in "+real)
	again := f.call(t, asOps, http.MethodGet, "/runs/"+row.ID, nil)
	assert.Equal(t, true, again.get("can_run"))

	args := f.call(t, asOps, http.MethodPost, "/runs", map[string]any{"folder": "infra", "dir": "app", "job": "@compose:up", "args": "ps"})
	assert.Equal(t, http.StatusBadRequest, args.Status)
	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Dir: "app", Job: "@compose:up"}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "@compose:up"}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "private", Job: "@git:status"}).Status)
}
