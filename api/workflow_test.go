package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func chain(pairs ...string) []api.StepRequest {
	steps := []api.StepRequest{}
	for i := 0; i+1 < len(pairs); i += 2 {
		steps = append(steps, api.StepRequest{Folder: pairs[i], Job: pairs[i+1]})
	}
	return steps
}

// createWorkflow makes a workflow as a user and returns its ID.
func (f *fixture) createWorkflow(t *testing.T, user, name string, steps []api.StepRequest) string {
	t.Helper()
	res := f.call(t, user, http.MethodPost, "/workflows", api.WorkflowRequest{Name: name, Steps: steps})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	return res.str("id")
}

// workflowRun waits for a workflow run to end and returns it.
func (f *fixture) workflowRun(t *testing.T, workflowID, runID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		res := f.call(t, asAdmin, http.MethodGet, "/workflows/"+workflowID+"/runs", nil)
		for _, item := range res.list("runs") {
			run := item.(map[string]any)
			if run["id"] == runID && run["state"] != model.StateRunning {
				return run
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("the workflow run did not end")
	return nil
}

func stepStates(run map[string]any) []string {
	states := []string{}
	for _, item := range run["steps"].([]any) {
		states = append(states, item.(map[string]any)["state"].(string))
	}
	return states
}

func TestWorkflow_Create(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "release", Steps: chain("infra", "plan", "infra/edge", "deploy")})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.True(t, model.ValidID(res.str("id")))
	assert.Equal(t, "ops", res.str("user_name"))
	assert.Len(t, res.list("steps"), 2)
	assert.Equal(t, true, res.get("can_run"))

	// Any user makes workflows; running them is another matter.
	viewer := f.call(t, asViewer, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "look", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusCreated, viewer.Status, string(viewer.Raw))
	assert.Equal(t, false, viewer.get("can_run"))

	cases := []struct {
		name string
		user string
		req  api.WorkflowRequest
	}{
		{"no name", asOps, api.WorkflowRequest{Steps: chain("infra", "plan")}},
		{"no steps", asOps, api.WorkflowRequest{Name: "empty"}},
		{"no such job", asOps, api.WorkflowRequest{Name: "x", Steps: chain("infra", "nope")}},
		{"no such folder", asOps, api.WorkflowRequest{Name: "x", Steps: chain("nowhere", "plan")}},
		{"a folder the user can not see", asLimited, api.WorkflowRequest{Name: "x", Steps: chain("private", "deploy")}},
	}
	for _, c := range cases {
		res := f.call(t, c.user, http.MethodPost, "/workflows", c.req)
		assert.Equal(t, http.StatusBadRequest, res.Status, c.name)
	}
}

func TestWorkflow_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.createWorkflow(t, asOps, "b infra", chain("infra", "plan"))
	f.createWorkflow(t, asAdmin, "a private", chain("infra", "plan", "private", "deploy"))

	admin := f.call(t, asAdmin, http.MethodGet, "/workflows", nil)
	assert.Len(t, admin.list("workflows"), 2)
	assert.Equal(t, "a private", admin.list("workflows")[0].(map[string]any)["name"])

	// A workflow with a step in a folder the user can not see is hidden.
	limited := f.call(t, asLimited, http.MethodGet, "/workflows", nil)
	assert.Len(t, limited.list("workflows"), 1)
	assert.Equal(t, "b infra", limited.list("workflows")[0].(map[string]any)["name"])
}

func TestWorkflow_Run(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ok := f.createWorkflow(t, asOps, "ok", chain("infra", "plan", "infra/edge", "deploy"))

	started := f.call(t, asOps, http.MethodPost, "/workflows/"+ok+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, ok, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, []string{"ok", "ok"}, stepStates(run))
	assert.Equal(t, "ops", run["user_name"])

	failing := f.createWorkflow(t, asOps, "failing", chain("infra", "plan", "infra", "fail", "infra", "plan"))
	started = f.call(t, asOps, http.MethodPost, "/workflows/"+failing+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status)
	run = f.workflowRun(t, failing, started.str("id"))
	assert.Equal(t, model.StateFailed, run["state"])
	assert.Equal(t, []string{"ok", "failed", "skipped"}, stepStates(run))

	// Running needs the right to run every step.
	viewer := f.call(t, asViewer, http.MethodPost, "/workflows/"+ok+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, viewer.Status)
	private := f.createWorkflow(t, asAdmin, "private", chain("infra", "plan", "private", "deploy"))
	denied := f.call(t, asOps, http.MethodPost, "/workflows/"+private+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)
	assert.Contains(t, denied.str("error"), "private / deploy")
}

func TestWorkflow_OneRunAtATime(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	slow := f.createWorkflow(t, asOps, "slow", chain("infra", "slow", "infra", "plan"))

	first := f.call(t, asOps, http.MethodPost, "/workflows/"+slow+"/runs", nil)
	assert.Equal(t, http.StatusCreated, first.Status)
	again := f.call(t, asOps, http.MethodPost, "/workflows/"+slow+"/runs", nil)
	assert.Equal(t, http.StatusConflict, again.Status)
	assert.Equal(t, first.str("id"), again.str("run_id"))

	shown := f.call(t, asOps, http.MethodGet, "/workflows/"+slow, nil)
	assert.Equal(t, first.str("id"), shown.str("running"))

	viewer := f.call(t, asViewer, http.MethodPost, "/workflow-runs/"+first.str("id")+"/stop", nil)
	assert.Equal(t, http.StatusForbidden, viewer.Status)
	stopped := f.call(t, asOps, http.MethodPost, "/workflow-runs/"+first.str("id")+"/stop", nil)
	assert.Equal(t, http.StatusNoContent, stopped.Status)
	run := f.workflowRun(t, slow, first.str("id"))
	assert.Equal(t, model.StateStopped, run["state"])
	assert.Equal(t, []string{"stopped", "skipped"}, stepStates(run))
}

func TestWorkflow_ChangeAndDelete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.createWorkflow(t, asOps, "mine", chain("infra", "plan"))

	other := f.call(t, asLimited, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "theirs", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusForbidden, other.Status)

	changed := f.call(t, asOps, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "mine v2", Steps: chain("infra", "plan", "infra", "fail")})
	assert.Equal(t, http.StatusOK, changed.Status, string(changed.Raw))
	assert.Equal(t, "mine v2", changed.str("name"))
	assert.Len(t, changed.list("steps"), 2)

	admin := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "by admin", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusOK, admin.Status)

	assert.Equal(t, http.StatusForbidden, f.call(t, asLimited, http.MethodDelete, "/workflows/"+id, nil).Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asOps, http.MethodDelete, "/workflows/"+id, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/workflows/"+id, nil).Status)
}
