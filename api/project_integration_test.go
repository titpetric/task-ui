//go:build integration

package api_test

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// localRepo makes a git repository with a Makefile, named website.git.
func localRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "website.git")
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\techo built\n"), 0o644))
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "."},
		{"-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		assert.NoError(t, err, string(out))
	}
	return dir
}

func TestProject_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/projects"},
		{http.MethodPost, "/projects"},
		{http.MethodGet, "/projects/infra"},
		{http.MethodPut, "/projects/infra"},
		{http.MethodDelete, "/projects/infra"},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

func TestProject_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodGet, "/projects", nil)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Len(t, res.list("entries"), 2)
	assert.Equal(t, "infra", res.list("entries")[0].(map[string]any)["id"])
	assert.Len(t, res.list("projects"), 0)
}

func TestProject_Clone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := localRepo(t)

	res := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + repo})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "website", res.str("project", "folder"))
	assert.Equal(t, true, res.get("runnable"))
	assert.FileExists(t, filepath.Join(f.root, "src", "website", "Makefile"))

	got := f.call(t, asAdmin, http.MethodGet, "/projects/website", nil)
	assert.Equal(t, http.StatusOK, got.Status)
	assert.Equal(t, "file://"+repo, got.str("source_url"))
	assert.Equal(t, false, got.get("shallow"), "recursive unless asked otherwise")

	no := false
	named := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + repo, ID: "site", Branch: "main", Recursive: &no})
	assert.Equal(t, http.StatusCreated, named.Status, string(named.Raw))
	assert.Equal(t, "main", named.str("project", "branch"))
	assert.Equal(t, true, f.call(t, asAdmin, http.MethodGet, "/projects/site", nil).get("shallow"))
	assert.DirExists(t, filepath.Join(f.root, "src", "site"))
}

func TestProject_CloneRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	bad := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "not a url"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	name := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "https://example.com/x.git", ID: "../x"})
	assert.Equal(t, http.StatusBadRequest, name.Status)

	internal := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + localRepo(t), ID: "internal"})
	assert.Equal(t, http.StatusBadRequest, internal.Status)
	assert.Contains(t, internal.str("error"), "task-ui's own names")
	under := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + localRepo(t), ID: "internal/tools"})
	assert.Equal(t, http.StatusBadRequest, under.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "internal"))

	missing := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + filepath.Join(t.TempDir(), "nothing.git")})
	assert.Equal(t, http.StatusBadGateway, missing.Status)
	assert.NotEqual(t, "", missing.str("error"))
	assert.NoDirExists(t, filepath.Join(f.root, "src", "nothing"))
}

func TestProject_Update(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	missing := f.call(t, asAdmin, http.MethodGet, "/projects/infra", nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)

	req := api.ProjectRequest{Title: "Infra", Description: "the servers", Allow: []string{"ops"}, Commands: map[string]string{"plan": "make plan"}, Sources: []string{"Makefile"}, JobAllow: map[string][]string{"plan": {"admin"}}}
	res := f.call(t, asAdmin, http.MethodPut, "/projects/infra", req)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, "Infra", res.str("title"))

	got := f.call(t, asAdmin, http.MethodGet, "/projects/infra", nil)
	assert.Equal(t, "the servers", got.str("description"))
	assert.Contains(t, got.str("allow"), "ops")
	assert.Contains(t, got.str("job_allow"), "admin")

	again := f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infrastructure"})
	assert.Equal(t, http.StatusOK, again.Status)
	assert.Equal(t, res.str("id"), again.str("id"))
	assert.Equal(t, "Infrastructure", again.str("title"))

	bad := f.call(t, asAdmin, http.MethodPut, "/projects/..", api.ProjectRequest{})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	internal := f.call(t, asAdmin, http.MethodPut, "/projects/internal", api.ProjectRequest{Title: "Mine"})
	assert.Equal(t, http.StatusBadRequest, internal.Status)
}

func TestProject_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := localRepo(t)

	cloned := f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: "file://" + repo})
	assert.Equal(t, http.StatusCreated, cloned.Status, string(cloned.Raw))

	res := f.call(t, asAdmin, http.MethodDelete, "/projects/website", nil)
	assert.Equal(t, http.StatusNoContent, res.Status)
	assert.NoDirExists(t, filepath.Join(f.root, "src", "website"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/projects/website", nil).Status)

	// A plain directory is not a checkout: the files stay.
	plain := f.call(t, asAdmin, http.MethodDelete, "/projects/infra", nil)
	assert.Equal(t, http.StatusNoContent, plain.Status)
	assert.DirExists(t, filepath.Join(f.root, "src", "infra"))

	bad := f.call(t, asAdmin, http.MethodDelete, "/projects/..", nil)
	assert.Equal(t, http.StatusBadRequest, bad.Status)
}

// A project is recursive unless made or set shallow: then the listing
// shows the project and its top-level folders alone.
func TestProject_Shallow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, false, res.get("shallow"))

	no := false
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra", Recursive: &no})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, true, res.get("shallow"))

	// A request without the field leaves it unchanged.
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra"})
	assert.Equal(t, true, res.get("shallow"))
	yes := true
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra", Recursive: &yes})
	assert.Equal(t, false, res.get("shallow"))
}

// A project shows in the sidebar unless set not to.
func TestProject_Sidebar(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, false, res.get("sidebar_hidden"))

	no := false
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra", Sidebar: &no})
	assert.Equal(t, true, res.get("sidebar_hidden"))
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra"})
	assert.Equal(t, true, res.get("sidebar_hidden")) // left out: unchanged
	yes := true
	res = f.call(t, asAdmin, http.MethodPut, "/projects/infra", api.ProjectRequest{Title: "Infra", Sidebar: &yes})
	assert.Equal(t, false, res.get("sidebar_hidden"))
}
