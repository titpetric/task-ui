//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A workflow runs as a user of its own: the creator unless an admin
// names another, a runner only themself (403), a name nobody holds a
// 404. Run starts the run as that user, who must be able to run every
// step as the caller must, a 403 naming the step and the user when they
// may not; an update changes the user, left out keeps it, and a user
// since deleted answers 409.
func TestWorkflow_RunAs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	own := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "own", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusCreated, own.Status, string(own.Raw))
	assert.Equal(t, "ops", own.str("run_as"))
	assert.Equal(t, "", own.str("run_as_user_id"), "left out, the owner runs it")
	assert.True(t, own.get("can_run").(bool))

	other := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "other", Steps: chain("infra", "plan"), User: asAdmin})
	assert.Equal(t, http.StatusForbidden, other.Status, string(other.Raw))
	assert.Contains(t, other.str("error"), "an admin picks another user")
	nobody := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "nobody", Steps: chain("infra", "plan"), User: "nobody"})
	assert.Equal(t, http.StatusNotFound, nobody.Status, string(nobody.Raw))

	asOpsRes := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "as ops", Steps: chain("infra", "plan"), User: asOps})
	assert.Equal(t, http.StatusCreated, asOpsRes.Status, string(asOpsRes.Raw))
	assert.Equal(t, "ops", asOpsRes.str("run_as"))
	assert.Equal(t, f.users[asOps].ID, asOpsRes.str("run_as_user_id"))
	assert.Equal(t, "admin", asOpsRes.str("user_name"), "the owner stays the creator")
	id := asOpsRes.str("id")
	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, id, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, "ops", run["user_name"], "the run is the run-as user's")

	// The run-as user's permissions bound a start besides the caller's:
	// ops may not run private / deploy, so the admin's start fails and
	// can_run says so.
	private := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "private as ops", Steps: chain("infra", "plan", "private", "deploy"), User: asOps})
	assert.Equal(t, http.StatusCreated, private.Status, string(private.Raw))
	assert.False(t, private.get("can_run").(bool))
	denied := f.call(t, asAdmin, http.MethodPost, "/workflows/"+private.str("id")+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status, string(denied.Raw))
	assert.Equal(t, "ops may not run private / deploy", denied.str("error"))

	// A prompt step asks Claude as the run-as user: its thread is theirs.
	miguel := &model.User{Name: "miguel", Role: model.RoleAdmin}
	assert.NoError(t, f.opts.Users.Create(ctx, miguel))
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `printf 'A: %s\n' "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}
	asks := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Steps: []api.StepRequest{{Kind: model.StepPrompt, Prompt: "what changed"}}, User: "miguel"})
	assert.Equal(t, http.StatusCreated, asks.Status, string(asks.Raw))
	assert.Equal(t, "miguel", asks.str("run_as"))
	started = f.call(t, asAdmin, http.MethodPost, "/workflows/"+asks.str("id")+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run = f.workflowRun(t, asks.str("id"), started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, "miguel", run["user_name"])
	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+started.str("id"), nil).list("notes")
	assert.Len(t, threads, 1)
	assert.Equal(t, "miguel", threads[0].(map[string]any)["user_name"], "the prompt went as the run-as user")

	// An update changes the user; one without keeps it; a runner may
	// name only themself.
	changed := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "as ops", Steps: chain("infra", "plan"), User: "miguel"})
	assert.Equal(t, http.StatusOK, changed.Status, string(changed.Raw))
	assert.Equal(t, "miguel", changed.str("run_as"))
	kept := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "as ops still", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusOK, kept.Status, string(kept.Raw))
	assert.Equal(t, "miguel", kept.str("run_as"))
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/workflows/"+own.str("id"), api.WorkflowRequest{Name: "own", Steps: chain("infra", "plan"), User: asAdmin}).Status)
	assert.Equal(t, http.StatusOK, f.call(t, asOps, http.MethodPut, "/workflows/"+own.str("id"), api.WorkflowRequest{Name: "own", Steps: chain("infra", "plan"), User: asOps}).Status)

	// The list carries run_as on every workflow; a user since deleted
	// leaves it empty, and a start answers 409.
	assert.NoError(t, f.opts.Users.Delete(ctx, miguel.ID))
	var listed struct {
		Workflows []api.WorkflowView `json:"workflows"`
	}
	assert.NoError(t, json.Unmarshal(f.call(t, asAdmin, http.MethodGet, "/workflows", nil).Raw, &listed))
	runsAs := map[string]string{}
	for _, wf := range listed.Workflows {
		runsAs[wf.Name] = wf.RunAs
	}
	assert.Equal(t, map[string]string{"own": "ops", "as ops still": "", "private as ops": "ops", "asks": ""}, runsAs)
	gone := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusConflict, gone.Status, string(gone.Raw))
	assert.Contains(t, gone.str("error"), "is gone")
}
