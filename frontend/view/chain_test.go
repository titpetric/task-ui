package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A chain names the folder once for the steps that follow in it, and a
// workflow step by its name.
func TestChain(t *testing.T) {
	t.Parallel()
	links := view.Chain([]view.Step{
		{Folder: "dockerfiles", Job: "build"},
		{Folder: "dockerfiles", Job: "push"},
		{Folder: "deploy/terraform", Job: "fmt"},
		{WorkflowID: "w1", WorkflowName: "release"},
		{Folder: "deploy/terraform", Job: "apply", State: "ok", RunID: "r1"},
	})
	texts, hrefs := []string{}, []string{}
	for _, l := range links {
		texts = append(texts, l.Text)
		hrefs = append(hrefs, l.Href)
	}
	assert.Equal(t, []string{"dockerfiles / build", "push", "deploy/terraform / fmt", "release", "deploy/terraform / apply"}, texts)
	assert.Equal(t, []string{"/f/dockerfiles/build", "/f/dockerfiles/push", "/f/deploy/terraform/fmt", "/workflows/w1", "/runs/r1"}, hrefs)
	assert.True(t, links[3].Workflow)
	assert.Equal(t, "ok", links[4].State)
}
