package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A step folds to its kind and one line: the job and its args, the
// workflow it runs, the first line of its prompt, or where its summary
// asks.
func TestStep(t *testing.T) {
	t.Parallel()
	job := view.Step{Folder: "infra", Job: "plan", Args: `["-target","db"]`}
	assert.Equal(t, model.StepJob, job.StepKind())
	assert.Equal(t, "Command", job.KindLabel())
	assert.Equal(t, "terminal", job.Icon())
	assert.Equal(t, "infra / plan -target db", job.Line())

	flow := view.Step{WorkflowID: "w1", WorkflowName: "deploy"}
	assert.Equal(t, "Workflow", flow.KindLabel())
	assert.Equal(t, "flow", flow.Icon())
	assert.Equal(t, "deploy", flow.Line())
	assert.Equal(t, "a workflow that is gone", view.Step{Kind: model.StepWorkflow, WorkflowID: "gone"}.Line())

	prompt := view.Step{Kind: model.StepPrompt, Folder: "_", Prompt: "  Fetch the weather.\nThen a table."}
	assert.Equal(t, "Prompt", prompt.KindLabel())
	assert.Equal(t, "prompt", prompt.Icon())
	assert.Equal(t, "Fetch the weather.", prompt.Line())

	summary := view.Step{Kind: model.StepSummary, Folder: "malta", Dir: "vhosts"}
	assert.Equal(t, "Summary", summary.KindLabel())
	assert.Equal(t, "summary", summary.Icon())
	assert.Equal(t, "of the steps before it, in malta/vhosts", summary.Line())
	assert.Equal(t, "the root", view.Step{Kind: model.StepSummary, Folder: "_"}.Place())
}

// A blank step of the editor: a command picks its folder, a prompt and
// a summary ask in the root.
func TestNewStep(t *testing.T) {
	t.Parallel()
	assert.Equal(t, view.Step{Kind: model.StepJob}, view.NewStep(model.StepJob))
	assert.Equal(t, view.Step{Kind: model.StepPrompt, Folder: "_"}, view.NewStep(model.StepPrompt))
	assert.Equal(t, "", view.NewStep(model.StepJob).Line())
	assert.Equal(t, "of the steps before it, in the root", view.NewStep(model.StepSummary).Line())
}
