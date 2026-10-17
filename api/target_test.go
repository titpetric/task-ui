package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// The target of a request is the folder and job the user may see; what a
// user may not see does not exist for them, and a job they may see but
// not run is refused with a reason.
func TestTarget(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	hidden := f.call(t, asLimited, http.MethodPost, "/runs", api.RunRequest{Folder: "private", Job: "deploy"})
	assert.Equal(t, http.StatusNotFound, hidden.Status)
	assert.Equal(t, "no such folder", hidden.str("error"))

	missing := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "nothing"})
	assert.Equal(t, http.StatusNotFound, missing.Status)
	assert.Equal(t, "no such job", missing.str("error"))

	denied := f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: "private", Job: "deploy"})
	assert.Equal(t, http.StatusForbidden, denied.Status)
	assert.Contains(t, denied.str("error"), "not run")

	// Admin only routes say so.
	admin := f.call(t, asViewer, http.MethodGet, "/users", nil)
	assert.Equal(t, http.StatusForbidden, admin.Status)
	assert.Equal(t, "admin role required", admin.str("error"))
}
