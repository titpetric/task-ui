package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The pipeline is the workflow's steps with the states of its latest
// run, when that run had the same steps; a run of an older chain lends
// no states.
func TestPipeline(t *testing.T) {
	t.Parallel()
	def := []view.Step{{Folder: "infra", Job: "plan"}, {WorkflowID: "w", WorkflowName: "release"}}
	last := []view.Step{{Folder: "infra", Job: "plan", State: "ok", RunID: "r1"}, {WorkflowID: "w", WorkflowName: "release", State: "running", WorkflowRunID: "wr"}}

	links := view.Pipeline(def, last)
	assert.Len(t, links, 2)
	assert.Equal(t, "ok", links[0].State)
	assert.Equal(t, "/runs/r1", links[0].Href)
	assert.Equal(t, "running", links[1].State)
	assert.Equal(t, "infra / plan", links[0].Text)

	older := []view.Step{{Folder: "infra", Job: "deploy", State: "ok", RunID: "r0"}}
	links = view.Pipeline(def, older)
	assert.Equal(t, "", links[0].State)
	assert.Equal(t, "/f/infra/plan", links[0].Href)

	links = view.Pipeline(def, nil)
	assert.Len(t, links, 2)
	assert.Equal(t, "", links[1].State)
}
