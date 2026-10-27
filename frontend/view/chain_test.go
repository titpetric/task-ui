package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A chain names the folder once for the steps that follow in it, and a
// workflow step by its name. A job of a project at the root links under
// /f/internal/, one of a workspace by its folder.
func TestChain(t *testing.T) {
	t.Parallel()
	links := view.Chain([]view.Step{
		{Folder: "dockerfiles", Job: "build"},
		{Folder: "dockerfiles", Job: "push"},
		{Folder: "deploy/terraform", Job: "fmt"},
		{WorkflowID: "w1", WorkflowName: "release"},
		{Folder: "deploy/terraform", Job: "apply", State: "ok", RunID: "r1"},
	}, map[string]int{"deploy": 0})
	texts, hrefs := []string{}, []string{}
	for _, l := range links {
		texts = append(texts, l.Text)
		hrefs = append(hrefs, l.Href)
	}
	assert.Equal(t, []string{"dockerfiles / build", "push", "deploy/terraform / fmt", "release", "deploy/terraform / apply"}, texts)
	assert.Equal(t, []string{"/f/internal/dockerfiles/build", "/f/internal/dockerfiles/push", "/f/deploy/terraform/fmt", "/workflows/w1", "/runs/r1"}, hrefs)
	assert.True(t, links[3].Workflow)
	assert.Equal(t, "ok", links[4].State)
}

// A prompt step reads as a prompt in its place, links to its folder's
// page, or to its run once it ran, and names its folder again for the
// job after it.
func TestChain_Prompt(t *testing.T) {
	t.Parallel()
	links := view.Chain([]view.Step{
		{Folder: "infra", Job: "plan"},
		{Folder: "infra", Dir: "docs", Job: "@claude-prompt", Prompt: "fetch the weather"},
		{Folder: "infra", Job: "apply"},
		{Folder: "internal", Job: "@claude-prompt", Prompt: "water today?", RunID: "r1"},
		{Job: "@claude-prompt", Prompt: "everywhere"},
	}, nil)
	assert.Equal(t, "prompt in infra/docs", links[1].Text)
	assert.Equal(t, "fetch the weather", links[1].Prompt)
	assert.Equal(t, "/f/internal/infra", links[1].Href)
	assert.Equal(t, "infra / apply", links[2].Text)
	assert.Equal(t, "prompt in the root", links[3].Text)
	assert.Equal(t, "/runs/r1", links[3].Href)
	assert.Equal(t, "prompt in the workspace root", links[4].Text)
}

// A link of a run's chain points at the run a step made, and one of a
// workflow step at the run it made of the workflow.
func TestLink(t *testing.T) {
	t.Parallel()
	links := view.Chain([]view.Step{
		{Folder: "infra", Job: "plan", State: "ok", RunID: "r1"},
		{WorkflowID: "w1", WorkflowName: "release", State: "running", WorkflowRunID: "wr1"},
	}, nil)
	var job, flow view.Link = links[0], links[1]
	assert.Equal(t, "r1", job.RunID)
	assert.False(t, job.Workflow)
	assert.True(t, flow.Workflow)
	assert.Equal(t, "w1", flow.WorkflowID)
	assert.Equal(t, "wr1", flow.WorkflowRunID)
	assert.Equal(t, "running", flow.State)
}
