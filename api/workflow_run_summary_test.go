package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A failed workflow run's summary comes from Claude once, over the report
// of its steps, and the API keeps it; a user who may not see every step
// gets none.
func TestWorkflowRunSummary(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	failing := f.createWorkflow(t, asAdmin, "failing", chain("infra", "plan", "infra", "fail", "infra", "plan"))
	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+failing+"/runs", nil)
	id := started.str("id")
	assert.Equal(t, model.StateFailed, f.workflowRun(t, failing, id)["state"])

	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workflow-runs/"+id+"/summary", nil).Status)
	f.call(t, asAdmin, http.MethodPut, "/integrations/claude", api.IntegrationRequest{Values: map[string]string{"token": "sk-1"}})
	written := f.call(t, asAdmin, http.MethodPost, "/workflow-runs/"+id+"/summary", map[string]any{})
	assert.Equal(t, http.StatusOK, written.Status, string(written.Raw))
	assert.Equal(t, "summary by sk-1 of # Workflow failing, run "+id, written.str("summary"))
	assert.Equal(t, false, written.Body["kept"])
	assert.Equal(t, true, f.call(t, asAdmin, http.MethodPost, "/workflow-runs/"+id+"/summary", map[string]any{}).Body["kept"])
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodGet, "/workflow-runs/"+id+"/summary", nil).Status)

	private := f.createWorkflow(t, asAdmin, "private", chain("private", "deploy"))
	hidden := f.call(t, asAdmin, http.MethodPost, "/workflows/"+private+"/runs", nil).str("id")
	f.workflowRun(t, private, hidden)
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodPost, "/workflow-runs/"+hidden+"/summary", map[string]any{}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPost, "/workflow-runs/"+model.NewID()+"/summary", map[string]any{}).Status)
}
