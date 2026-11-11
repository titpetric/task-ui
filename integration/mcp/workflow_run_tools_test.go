package mcp

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run's steps are named by what they did, a check by the markers it
// looked for, and a step that has not run yet says so.
func TestShowFlowRun(t *testing.T) {
	t.Parallel()
	view := api.WorkflowRunView{
		WorkflowRun: model.WorkflowRun{ID: "r1", Name: "refine", UserName: "ops", State: model.StateOK, Answer: "done"},
		Steps: []model.WorkflowRunStep{
			{Kind: model.StepPrompt, Folder: model.Default, Prompt: "refine", State: model.StateOK},
			{Kind: model.StepCheck, Markers: "Next:", State: model.StateOK, Answer: `found "Next:" in step 1`},
			{Kind: model.StepJob, Folder: "infra", Job: "plan"},
		},
	}
	text := showFlowRun(view, nil)
	for _, want := range []string{"workflow run r1: refine by ops", "1. prompt in ", "2. check for Next:: ok", "3. infra/plan: not run", "answer:\ndone"} {
		assert.True(t, strings.Contains(text, want), want+" in\n"+text)
	}
}
