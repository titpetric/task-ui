package api_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A workflow takes a prompt and where it runs, and steps that are off;
// one with no enabled step and no prompt answers 400, a prompt takes
// prompt/edit and a place the user sees.
func TestWorkflow_PromptAndToggles(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	off := false

	refused := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "idle", Steps: []api.StepRequest{{Folder: "infra", Job: "plan", Enabled: &off}}})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
	assert.Contains(t, refused.str("error"), "an enabled step or a prompt")
	empty := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "idle"})
	assert.Equal(t, http.StatusBadRequest, empty.Status)

	ops := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Prompt: "say ok"})
	assert.Equal(t, http.StatusForbidden, ops.Status, string(ops.Raw))
	gone := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Prompt: "say ok", PromptFolder: "nowhere"})
	assert.Equal(t, http.StatusBadRequest, gone.Status, string(gone.Raw))

	made := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{
		Name: "summed", Prompt: " summarize, change nothing ", PromptFolder: "infra",
		Steps: []api.StepRequest{{Folder: "infra", Job: "plan"}, {Folder: "infra", Job: "fail", Enabled: &off}},
	})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	assert.Equal(t, "summarize, change nothing", made.str("prompt"))
	assert.Equal(t, "infra", made.str("prompt_folder"))
	steps := made.list("steps")
	assert.Equal(t, false, steps[0].(map[string]any)["disabled"])
	assert.Equal(t, true, steps[1].(map[string]any)["disabled"])

	// The root is the empty place; a workflow without a prompt keeps none.
	id := made.str("id")
	root := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "summed", Prompt: "say ok", PromptFolder: "_", Steps: chain("infra", "plan")})
	assert.Equal(t, http.StatusOK, root.Status, string(root.Raw))
	assert.Equal(t, "", root.str("prompt_folder"))
	plain := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "summed", PromptFolder: "infra", Steps: chain("infra", "plan")})
	assert.Equal(t, "", plain.str("prompt"))
	assert.Equal(t, "", plain.str("prompt_folder"))
}

// A run of a workflow with a prompt runs its enabled steps, skips the
// rest, then puts the prompt to Claude with the run's input as the user
// who ran it; the thread lists by the workflow run, and the run ends
// with Claude's reply. A prompt-only workflow is the prompt alone.
// Running one takes prompt/edit.
func TestWorkflow_RunPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `{ printf 'Q: %s\n' "$TASK_UI_PROMPT"; cat "$TASK_UI_PROMPT_WORKFLOW"; } > "$TASK_UI_PROMPT_REPLY"`}
	off := false
	made := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{
		Name: "summed", Prompt: "summarize, change nothing",
		Steps: []api.StepRequest{{Folder: "infra", Job: "plan"}, {Folder: "infra", Job: "fail", Enabled: &off}},
	})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	id := made.str("id")

	refused := f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, refused.Status, string(refused.Raw))

	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", api.WorkflowRunRequest{Input: "for the 1.2 tag"})
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	runID := started.str("id")
	run := f.workflowRun(t, id, runID)
	assert.Equal(t, model.StateOK, run["state"])
	assert.Equal(t, []string{"ok", "skipped"}, stepStates(run))

	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+runID, nil).list("notes")
	assert.Len(t, threads, 1)
	thread := threads[0].(map[string]any)
	assert.Equal(t, "summarize, change nothing\n\nInput:\nfor the 1.2 tag", thread["body"])
	assert.Equal(t, runID, thread["about_workflow_run_id"])
	assert.Equal(t, asAdmin, thread["user_name"])
	assert.Equal(t, model.NoteDone, thread["state"])
	answer := thread["replies"].([]any)[0].(map[string]any)["body"].(string)
	assert.Contains(t, answer, "Q: summarize, change nothing\n\nInput:\nfor the 1.2 tag")
	assert.Contains(t, answer, "## Step 1, infra / plan: ok, exit code 0")
	assert.Contains(t, answer, "planned")
	assert.Contains(t, answer, "## Step 2, infra / fail: skipped")
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run=nothing", nil).Status)

	alone := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "alone", Prompt: "say ok, change nothing"})
	assert.Equal(t, http.StatusCreated, alone.Status, string(alone.Raw))
	aloneID := alone.str("id")
	again := f.call(t, asAdmin, http.MethodPost, "/workflows/"+aloneID+"/runs", nil)
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))
	assert.Equal(t, model.StateOK, f.workflowRun(t, aloneID, again.str("id"))["state"])
	said := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+again.str("id"), nil).list("notes")
	assert.Len(t, said, 1)
	assert.Equal(t, "say ok, change nothing", said[0].(map[string]any)["body"])
}

// Stop on a workflow run that is at its prompt stops the prompt's run,
// and the workflow run ends stopped.
func TestWorkflow_StopPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.opts.Board.AgentCommand = []string{"sh", "-c", `sleep 30`}
	made := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "waits", Prompt: "say ok"})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	id := made.str("id")
	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
	runID := started.str("id")
	assert.Eventually(t, func() bool {
		threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+runID, nil).list("notes")
		return len(threads) == 1 && threads[0].(map[string]any)["state"] == model.NoteWorking
	}, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodPost, "/workflow-runs/"+runID+"/stop", nil).Status)
	run := f.workflowRun(t, id, runID)
	assert.Equal(t, model.StateStopped, run["state"])
	assert.Equal(t, "stopped at the prompt", run["message"])
	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+runID, nil).list("notes")
	assert.Equal(t, model.NoteFailed, threads[0].(map[string]any)["state"])
}
