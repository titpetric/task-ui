//go:build integration

package api_test

import (
	"context"
	"encoding/json"
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
		{"a shell", asAdmin, api.WorkflowRequest{Name: "x", Steps: chain("infra", "plan", "infra", "@shell")}},
	}
	for _, c := range cases {
		res := f.call(t, c.user, http.MethodPost, "/workflows", c.req)
		assert.Equal(t, http.StatusBadRequest, res.Status, c.name)
	}
}

// A step picks from the catalogue the editor lists, every folder with a
// job, so a folder the settings hide from the listings saves and runs;
// the folders a user may see still bound it.
func TestWorkflow_HiddenFolder(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	listed := f.call(t, asOps, http.MethodGet, "/folders", nil)
	assert.NotContains(t, string(listed.Raw), `"infra/hidden"`, "the listings hide it")

	res := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "hidden", Steps: chain("infra/hidden", "deploy")})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, true, res.get("can_run"))
	id := res.str("id")

	limited := f.call(t, asLimited, http.MethodGet, "/workflows/"+id, nil)
	assert.Equal(t, http.StatusNotFound, limited.Status, "the user's folders name infra alone")
	refused := f.call(t, asLimited, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "x", Steps: chain("infra/hidden", "deploy")})
	assert.Equal(t, http.StatusBadRequest, refused.Status)
	assert.Equal(t, "step 1: infra/hidden / deploy: Internal has no folder infra/hidden", refused.str("error"))

	denied := f.call(t, asViewer, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)

	started := f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, id, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
}

// The error of a step whose folder is not there names the step by its
// folder relative to its workspace and the job, and the workspace by
// its title.
func TestWorkflow_FolderError(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(context.Background()))
	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: "malta", Title: "Malta"}))

	for req, want := range map[string]string{
		"nowhere":                "step 1: nowhere / plan: Internal has no folder nowhere",
		"malta/deploy/terraform": "step 1: deploy/terraform / plan: Malta has no folder deploy/terraform",
		"malta":                  "step 1: the root / plan: Malta has no jobs at its root",
		"private":                "step 1: private / plan: Internal has no folder private",
	} {
		user := asOps
		if req == "private" {
			user = asLimited
		}
		res := f.call(t, user, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "x", Steps: chain(req, "plan")})
		assert.Equal(t, http.StatusBadRequest, res.Status, req)
		assert.Equal(t, want, res.str("error"), req)
	}
	res := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "x", Steps: chain("infra", "plan", "infra", "nope")})
	assert.Equal(t, "step 2: infra / nope: no such job", res.str("error"))
}

func TestWorkflow_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.createWorkflow(t, asOps, "b infra", chain("infra", "plan"))
	f.createWorkflow(t, asAdmin, "a private", chain("infra", "plan", "private", "deploy"))

	admin := f.call(t, asAdmin, http.MethodGet, "/workflows", nil)
	assert.Len(t, admin.list("workflows"), 2)
	assert.Equal(t, "a private", admin.list("workflows")[0].(map[string]any)["name"])

	// The listing hides a workflow with a step in a folder the user can not see.
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

// The editor saves the whole list in its new order: a step moved up
// reads back at its new position, with its kind and fields.
func TestWorkflow_Reorder(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	summary := api.StepRequest{Kind: model.StepSummary, Folder: "infra"}
	id := f.createWorkflow(t, asAdmin, "order", append(chain("infra", "plan", "infra", "fail"), summary))
	moved := append([]api.StepRequest{summary}, chain("infra", "fail", "infra", "plan")...)
	res := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "order", Steps: moved})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.WorkflowView
	assert.NoError(t, json.Unmarshal(f.call(t, asAdmin, http.MethodGet, "/workflows/"+id, nil).Raw, &view))
	assert.Len(t, view.Steps, 3)
	assert.Equal(t, model.StepSummary, view.Steps[0].Kind)
	assert.Equal(t, "infra", view.Steps[0].Folder)
	assert.Equal(t, int64(0), view.Steps[0].Position)
	assert.Equal(t, "fail", view.Steps[1].Job)
	assert.Equal(t, "plan", view.Steps[2].Job)
	assert.Equal(t, int64(2), view.Steps[2].Position)
}

// A step may name another workflow: the user sees it when they see
// every step of it, and runs it when they may run them all.
func TestWorkflow_Nested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	inner := f.createWorkflow(t, asOps, "inner", chain("infra", "plan"))

	res := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "outer", Steps: []api.StepRequest{{Folder: "infra", Job: "plan"}, {Workflow: inner}}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	outer := res.str("id")
	assert.Equal(t, inner, res.list("steps")[1].(map[string]any)["workflow_ref"])
	assert.Equal(t, model.StepWorkflow, res.list("steps")[1].(map[string]any)["kind"])

	missing := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "gone", Steps: []api.StepRequest{{Workflow: "nope"}}})
	assert.Equal(t, http.StatusBadRequest, missing.Status)
	loop := f.call(t, asOps, http.MethodPut, "/workflows/"+inner, api.WorkflowRequest{Name: "inner", Steps: []api.StepRequest{{Workflow: outer}}})
	assert.Equal(t, http.StatusBadRequest, loop.Status)
	assert.Contains(t, loop.str("error"), "loop")

	started := f.call(t, asOps, http.MethodPost, "/workflows/"+outer+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, outer, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, []string{"ok", "ok"}, stepStates(run))
	nested := run["steps"].([]any)[1].(map[string]any)
	assert.True(t, nested["workflow_run_ref"] != "", "the run of inner")

	// The API hides a workflow around a folder the user can not see, and
	// one around a job they may not run is not theirs to run.
	private := f.createWorkflow(t, asAdmin, "private", chain("private", "deploy"))
	wrap := f.createWorkflow(t, asAdmin, "wrap", []api.StepRequest{{Workflow: private}})
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/workflows/"+wrap, nil).Status)
	seen := f.call(t, asOps, http.MethodGet, "/workflows/"+wrap, nil)
	assert.Equal(t, http.StatusOK, seen.Status)
	assert.Equal(t, false, seen.get("can_run"))
	denied := f.call(t, asOps, http.MethodPost, "/workflows/"+wrap+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)
	assert.Contains(t, denied.str("error"), "private / deploy")

	// A workflow another one runs stays.
	used := f.call(t, asOps, http.MethodDelete, "/workflows/"+inner, nil)
	assert.Equal(t, http.StatusConflict, used.Status)
	assert.Contains(t, used.str("error"), "outer")
}

// The runs of every workflow the user sees, newest first.
func TestWorkflow_RecentRuns(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ok := f.createWorkflow(t, asOps, "ok", chain("infra", "plan"))
	private := f.createWorkflow(t, asAdmin, "private", chain("private", "deploy"))
	for _, id := range []string{ok, private} {
		started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
		assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
		f.workflowRun(t, id, started.str("id"))
	}

	admin := f.call(t, asAdmin, http.MethodGet, "/workflow-runs?limit=10", nil)
	assert.Equal(t, http.StatusOK, admin.Status, string(admin.Raw))
	assert.Len(t, admin.list("runs"), 2)
	assert.Equal(t, "private", admin.list("runs")[0].(map[string]any)["name"])

	limited := f.call(t, asLimited, http.MethodGet, "/workflow-runs", nil)
	runs := limited.list("runs")
	assert.Len(t, runs, 1)
	assert.Equal(t, "ok", runs[0].(map[string]any)["name"])
	assert.Equal(t, ok, runs[0].(map[string]any)["workflow_id"])
}

// A workflow shows in the sidebar unless made or set not to.
func TestWorkflow_Sidebar(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	no, yes := false, true

	res := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "part", Steps: chain("infra", "plan"), Sidebar: &no})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, true, res.get("sidebar_hidden"))
	id := res.str("id")

	res = f.call(t, asOps, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "part", Steps: chain("infra", "plan")})
	assert.Equal(t, true, res.get("sidebar_hidden")) // left out: unchanged
	res = f.call(t, asOps, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "part", Steps: chain("infra", "plan"), Sidebar: &yes})
	assert.Equal(t, false, res.get("sidebar_hidden"))

	plain := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "whole", Steps: chain("infra", "plan")})
	assert.Equal(t, false, plain.get("sidebar_hidden"))
}

// A workflow decodes into its view: the row, the steps, and what the
// user asking may do; its runs into theirs, each with its steps.
func TestWorkflowView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.createWorkflow(t, asOps, "ship", chain("infra", "plan", "infra/edge", "deploy"))

	res := f.call(t, asOps, http.MethodGet, "/workflows/"+id, nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.WorkflowView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "ship", view.Name)
	assert.Len(t, view.Steps, 2)
	assert.Equal(t, "infra/edge", view.Steps[1].Folder)
	assert.True(t, view.CanRun)
	assert.True(t, view.CanChange)

	var viewer api.WorkflowView
	assert.NoError(t, json.Unmarshal(f.call(t, asViewer, http.MethodGet, "/workflows/"+id, nil).Raw, &viewer))
	assert.False(t, viewer.CanRun)
	assert.False(t, viewer.CanChange)
}

func TestWorkflowRunView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.createWorkflow(t, asOps, "ship", chain("infra", "plan"))
	started := f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	f.workflowRun(t, id, started.str("id"))

	var body struct {
		Runs []api.WorkflowRunView `json:"runs"`
	}
	assert.NoError(t, json.Unmarshal(f.call(t, asOps, http.MethodGet, "/workflows/"+id+"/runs", nil).Raw, &body))
	assert.Len(t, body.Runs, 1)
	assert.Equal(t, started.str("id"), body.Runs[0].ID)
	assert.Equal(t, model.StateOK, body.Runs[0].State)
	assert.Len(t, body.Runs[0].Steps, 1)
	assert.Equal(t, "plan", body.Runs[0].Steps[0].Job)
}
