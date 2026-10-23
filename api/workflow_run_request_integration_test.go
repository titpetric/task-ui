//go:build integration

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

// A workflow takes steps of every kind and steps that are off; one with
// no enabled step answers 400, a prompt or summary takes prompt/edit
// and a place the user sees. A final prompt of a request is its last
// prompt step, in prompt_folder, and the workflow shows no prompt.
func TestWorkflow_PromptAndToggles(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	off := false

	refused := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "idle", Steps: []api.StepRequest{{Folder: "infra", Job: "plan", Enabled: &off}}})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
	assert.Contains(t, refused.str("error"), "a workflow needs an enabled step")
	empty := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "idle"})
	assert.Equal(t, http.StatusBadRequest, empty.Status)

	ops := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Prompt: "say ok"})
	assert.Equal(t, http.StatusForbidden, ops.Status, string(ops.Raw))
	opsSummary := f.call(t, asOps, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Steps: []api.StepRequest{{Kind: model.StepSummary}}})
	assert.Equal(t, http.StatusForbidden, opsSummary.Status, string(opsSummary.Raw))
	gone := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "asks", Prompt: "say ok", PromptFolder: "nowhere"})
	assert.Equal(t, http.StatusBadRequest, gone.Status, string(gone.Raw))

	made := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{
		Name: "summed", Description: " the plan, summed ", Prompt: " summarize, change nothing ", PromptFolder: "infra",
		Steps: []api.StepRequest{{Folder: "infra", Job: "plan"}, {Folder: "infra", Job: "fail", Enabled: &off}},
	})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	assert.Equal(t, "", made.str("prompt"))
	assert.Equal(t, "", made.str("prompt_folder"))
	assert.Equal(t, "the plan, summed", made.str("description"))
	steps := made.list("steps")
	assert.Len(t, steps, 3)
	assert.Equal(t, false, steps[0].(map[string]any)["disabled"])
	assert.Equal(t, true, steps[1].(map[string]any)["disabled"])
	last := steps[2].(map[string]any)
	assert.Equal(t, model.StepPrompt, last["kind"])
	assert.Equal(t, "summarize, change nothing", last["prompt"])
	assert.Equal(t, "infra", last["folder"])

	// A summary step asks in the root without a folder, and takes no
	// prompt; a prompt step needs one.
	id := made.str("id")
	summed := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "summed", Steps: append(chain("infra", "plan"), api.StepRequest{Kind: model.StepSummary})})
	assert.Equal(t, http.StatusOK, summed.Status, string(summed.Raw))
	summary := summed.list("steps")[1].(map[string]any)
	assert.Equal(t, model.StepSummary, summary["kind"])
	assert.Equal(t, model.Default, summary["folder"])
	assert.Equal(t, model.PromptJob, summary["job"])
	for _, bad := range []api.StepRequest{
		{Kind: model.StepSummary, Prompt: "say ok"},
		{Kind: model.StepPrompt},
		{Kind: model.StepJob, Folder: "infra"},
		{Kind: "other", Folder: "infra", Job: "plan"},
		{Kind: model.StepPrompt, Prompt: "say ok", Job: "plan"},
	} {
		res := f.call(t, asAdmin, http.MethodPut, "/workflows/"+id, api.WorkflowRequest{Name: "summed", Steps: append(chain("infra", "plan"), bad)})
		assert.Equal(t, http.StatusBadRequest, res.Status, string(res.Raw))
		assert.Contains(t, res.str("error"), "step 2")
	}
}

// A run of a workflow with a last prompt step runs its enabled steps,
// skips the rest, then puts the prompt to Claude with the run's input as
// the user who ran it; the thread lists by the workflow run, and the run
// ends with Claude's reply. A workflow of a prompt step alone is that.
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
	assert.Equal(t, []string{"ok", "skipped", "ok"}, stepStates(run))

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
	assert.Len(t, alone.list("steps"), 1)
	runs := f.call(t, asAdmin, http.MethodGet, "/workflows/"+aloneID+"/runs", nil).list("runs")
	assert.Len(t, runs[0].(map[string]any)["steps"].([]any), 1)
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
	assert.Equal(t, "stopped at the prompt in _", run["message"])
	threads := f.call(t, asAdmin, http.MethodGet, "/prompts?workflow_run="+runID, nil).list("notes")
	assert.Equal(t, model.NoteFailed, threads[0].(map[string]any)["state"])
}
