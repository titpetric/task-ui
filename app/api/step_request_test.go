package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A step names its kind, or its fields say it: a prompt with a prompt,
// a workflow with a workflow, else a job. A summary says so.
func TestStepRequest_StepKind(t *testing.T) {
	t.Parallel()
	assert.Equal(t, model.StepJob, api.StepRequest{Folder: "infra", Job: "plan"}.StepKind())
	assert.Equal(t, model.StepWorkflow, api.StepRequest{Workflow: "w1"}.StepKind())
	assert.Equal(t, model.StepPrompt, api.StepRequest{Prompt: "say ok"}.StepKind())
	assert.Equal(t, model.StepSummary, api.StepRequest{Kind: model.StepSummary, Folder: "_"}.StepKind())
	assert.Equal(t, "nonsense", api.StepRequest{Kind: "nonsense"}.StepKind(), "the API refuses it by name")

	var step api.StepRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"kind":"summary","folder":"_","dir":"app"}`), &step))
	assert.Equal(t, api.StepRequest{Kind: model.StepSummary, Folder: "_", Dir: "app"}, step)
}
