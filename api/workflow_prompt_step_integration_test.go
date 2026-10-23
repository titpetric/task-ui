//go:build integration

package api_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A prompt step asks Claude in its place, the root by default, with the
// steps before it; its answer is its output for the final prompt. A
// prompt step takes prompt/edit, a place the user sees, and no job.
func TestWorkflow_PromptStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `{ printf 'Q: %s\n' "$TASK_UI_PROMPT"; cat "$TASK_UI_PROMPT_WORKFLOW"; } > "$TASK_UI_PROMPT_REPLY"`}
	ask := api.StepRequest{Prompt: " fetch the weather "}

	refused := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "weather", Steps: []api.StepRequest{ask}})
	assert.Equal(t, http.StatusForbidden, refused.Status, string(refused.Raw))
	gone := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "weather", Steps: []api.StepRequest{{Prompt: "x", Folder: "nowhere"}}})
	assert.Equal(t, http.StatusBadRequest, gone.Status, string(gone.Raw))
	both := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "weather", Steps: []api.StepRequest{{Prompt: "x", Folder: "infra", Job: "plan"}}})
	assert.Equal(t, http.StatusBadRequest, both.Status, string(both.Raw))

	made := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{
		Name: "weather", Prompt: "water today?",
		Steps: []api.StepRequest{ask, {Folder: "infra", Job: "plan"}},
	})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	step := made.list("steps")[0].(map[string]any)
	assert.Equal(t, "fetch the weather", step["prompt"])
	assert.Equal(t, model.Default, step["folder"])
	assert.Equal(t, model.PromptJob, step["job"])
	id := made.str("id")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil).Status)

	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", api.WorkflowRunRequest{Input: "porec"})
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	runID := started.str("id")
	run := f.workflowRun(t, id, runID)
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, []string{"ok", "ok"}, stepStates(run))
	first := run["steps"].([]any)[0].(map[string]any)
	assert.Equal(t, "fetch the weather", first["prompt"])
	assert.True(t, first["run_id"] != "")
	assert.Contains(t, first["answer"].(string), "Q: fetch the weather\n\nInput:\nporec")

	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+runID, nil).list("notes")
	assert.Len(t, threads, 2)
	final := threads[0].(map[string]any)
	assert.Equal(t, "water today?\n\nInput:\nporec", final["body"])
	answer := final["replies"].([]any)[0].(map[string]any)["body"].(string)
	assert.Contains(t, answer, `## Step 1, a prompt to Claude in _, "fetch the weather": ok`)
	assert.Contains(t, answer, "Its answer:")
	assert.Contains(t, answer, "## Step 2, infra / plan: ok")
}
