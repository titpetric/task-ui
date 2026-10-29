//go:build integration

package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A script step runs its command in a folder, with shell, and a
// workflow runs over a workspace the user sees, without workflow steps;
// left out on an update, runs_over stays.
func TestWorkflow_ScriptRunsOver(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	script := []api.StepRequest{{Kind: model.StepJob, Folder: "infra", Job: folder.ScriptID, Args: []string{"echo scripted"}}}

	denied := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "denied", Steps: script})
	assert.Equal(t, http.StatusForbidden, denied.Status, string(denied.Raw))
	assert.Contains(t, denied.str("error"), "shell")

	empty := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "empty", Steps: []api.StepRequest{{Folder: "infra", Job: folder.ScriptID}}})
	assert.Equal(t, http.StatusBadRequest, empty.Status, string(empty.Raw))
	assert.Contains(t, empty.str("error"), "needs a command")

	nowhere := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "nowhere", Steps: script, RunsOver: ptr("nowhere")})
	assert.Equal(t, http.StatusBadRequest, nowhere.Status, string(nowhere.Raw))
	assert.Contains(t, nowhere.str("error"), "no workspace nowhere")

	created := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "maintain", Steps: script, RunsOver: ptr(model.Default)})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	assert.Equal(t, model.Default, created.str("runs_over"))
	id := created.str("id")

	nested := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "maintain", Steps: []api.StepRequest{{Kind: model.StepWorkflow, Workflow: id}}})
	assert.Equal(t, http.StatusBadRequest, nested.Status, string(nested.Raw))
	assert.Contains(t, nested.str("error"), "workflow step")

	kept := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "maintain", Steps: script})
	assert.Equal(t, http.StatusOK, kept.Status, string(kept.Raw))
	assert.Equal(t, model.Default, kept.str("runs_over"), "left out, it stays")

	cleared := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "maintain", Steps: script, RunsOver: ptr("")})
	assert.Equal(t, http.StatusOK, cleared.Status, string(cleared.Raw))
	assert.Equal(t, "", cleared.str("runs_over"))

	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, id, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"], run["message"])
}
