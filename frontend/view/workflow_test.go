package view_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/permission"
)

// A workflow as a page shows it: the row, its steps drawn as a chain,
// and what the user asking may do with it.
func TestWorkflow(t *testing.T) {
	t.Parallel()
	w := view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "release"},
		Steps:    []view.Step{{Folder: "dockerfiles", Job: "build"}, {Folder: "dockerfiles", Job: "push"}},
		CanRun:   true,
	}
	assert.Equal(t, "release", w.Name)
	assert.Equal(t, "push", view.Chain(w.Steps, nil)[1].Text)
	assert.True(t, w.CanRun)
	assert.False(t, w.CanChange)
	assert.False(t, w.Prompted())
	assert.Equal(t, "_", w.PromptPlace())

	w.Steps = append(w.Steps, view.Step{Kind: model.StepSummary, Folder: "infra", Dir: "app"})
	w.Steps[1].Disabled = true
	assert.True(t, w.Prompted())
	assert.Equal(t, "infra/app", w.PromptPlace())
	assert.True(t, view.Chain(w.Steps, nil)[1].Disabled)
	assert.True(t, view.Chain(w.Steps, nil)[2].Summary)
	assert.Equal(t, "summary in infra/app", view.Chain(w.Steps, nil)[2].Text)
	w.Steps[2].Disabled = true
	assert.False(t, w.Prompted(), "a step that is off asks nothing")
	box := view.PromptOfWorkflowRun("wr1", "infra/app", "infra/app/")
	assert.True(t, box.Watch)
	assert.Equal(t, "infra/app", box.Place)
	assert.Equal(t, "wr1", box.WorkflowRun)
}

// A run of a workflow carries the states of its steps.
func TestWorkflowRun(t *testing.T) {
	t.Parallel()
	r := view.WorkflowRun{
		WorkflowRun: model.WorkflowRun{ID: "wr1", State: model.StateFailed},
		Steps:       []view.Step{{Folder: "infra", Job: "plan", State: "ok"}, {Folder: "infra", Job: "apply", State: "failed"}},
	}
	assert.Equal(t, "wr1", r.ID)
	assert.Equal(t, "failed", view.Chain(r.Steps, nil)[1].State)
}

func TestWorkflowRun_Took(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Second)
	run := view.WorkflowRun{WorkflowRun: model.WorkflowRun{StartedAt: &start, FinishedAt: &end}}
	assert.Equal(t, 90*time.Second, run.Took())
}

// The editor draws each step as a card of its kind on the rail, with
// its [...] menu, and Add step lists the kinds; the prompt and summary
// kinds take prompt/edit.
func TestView_RenderWorkflowEditor(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflow = &view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "Porec watering", Description: "Water the plants or not."},
		Steps: []view.Step{
			{Kind: model.StepJob, Folder: "infra", Job: "plan"},
			{Kind: model.StepPrompt, Folder: "_", Prompt: "Fetch the weather for Porec.", Disabled: true},
			{Kind: model.StepSummary, Folder: "infra", Dir: "app"},
		},
		CanChange: true,
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	assert.Contains(t, out, `data-kind="job"`)
	assert.Contains(t, out, `class="wf-step kind-prompt off"`)
	assert.Contains(t, out, `data-kind="summary"`)
	assert.Contains(t, out, ">infra / plan</span>")
	assert.Contains(t, out, ">Fetch the weather for Porec.</span>")
	assert.Contains(t, out, ">of the steps before it, in infra/app</span>")
	assert.Contains(t, out, note.WorkflowSummaryQuestion)
	assert.Contains(t, out, `data-step-move="up"`)
	assert.Contains(t, out, `data-step-duplicate="1"`)
	assert.Contains(t, out, `<span>Enable</span>`, "the step that is off turns on")
	assert.Contains(t, out, `class="more-item danger" role="menuitem" data-step-remove="1"`)
	assert.Contains(t, out, `data-step-add="summary"`)
	assert.Contains(t, out, `data-step-template="prompt"`)
	assert.Contains(t, out, "Water the plants or not.")

	runner := &model.User{Name: "ops", Role: model.RoleRunner}
	p.User, p.Permissions = runner, permission.Set{User: runner, Role: permission.Policy{"workflow/edit": permission.Allow}}
	p.Workflow.Steps = nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_new", p))
	out = buf.String()
	assert.Contains(t, out, `data-step-add="job"`)
	assert.NotContains(t, out, `data-step-add="prompt"`, "no prompt/edit")
	assert.Contains(t, out, "No steps yet. Start with a command.")
}
