package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A finished run's summary comes from Claude once, signed in as the user
// asking; every later ask, and GET, returns the one kept.
func TestRunSummary(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	id := f.start(t, asAdmin, "infra", "plan")
	f.finished(t, id)

	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/runs/"+id+"/summary", nil).Status)
	unsigned := f.call(t, asAdmin, http.MethodPost, "/runs/"+id+"/summary", map[string]any{})
	assert.Equal(t, http.StatusBadGateway, unsigned.Status)
	assert.Contains(t, unsigned.str("error"), "no token")

	f.call(t, asAdmin, http.MethodPut, "/integrations/claude", api.IntegrationRequest{Values: map[string]string{"token": "sk-1"}})
	written := f.call(t, asAdmin, http.MethodPost, "/runs/"+id+"/summary", map[string]any{})
	assert.Equal(t, http.StatusOK, written.Status, string(written.Raw))
	assert.Equal(t, "summary by sk-1 of # infra / plan, run "+id, written.str("summary"))
	assert.Equal(t, false, written.Body["kept"])
	assert.Contains(t, written.str("html"), "<p>summary by sk-1")

	kept := f.call(t, asAdmin, http.MethodPost, "/runs/"+id+"/summary", map[string]any{})
	assert.Equal(t, true, kept.Body["kept"])
	assert.Equal(t, written.str("summary"), kept.str("summary"))
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodGet, "/runs/"+id+"/summary", nil).Status)
	// limited sees infra, and gets the kept one; a run of a folder they do
	// not see is not there.
	assert.Equal(t, true, f.call(t, asLimited, http.MethodPost, "/runs/"+id+"/summary", map[string]any{}).Body["kept"])
	hidden := f.start(t, asAdmin, "private", "deploy")
	f.finished(t, hidden)
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodPost, "/runs/"+hidden+"/summary", map[string]any{}).Status)
}
