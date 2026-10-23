package workflow_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workflow"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func job(id, script string) *folder.Job {
	j := &folder.Job{Folder: "infra", ID: id, Command: []string{"sh", "-c", script}}
	j.Name = id
	return j
}

// A step is a job of a folder, or a workflow in its place.
func TestStep(t *testing.T) {
	t.Parallel()
	j := job("build", "echo built")
	s := workflow.Step{Folder: &folder.Folder{ID: "infra"}, Job: j}
	assert.Equal(t, "build", s.Job.ID)
	assert.True(t, workflow.Step{Workflow: &model.Workflow{ID: "w1"}}.Job == nil)
}

// Anything that finds a folder by ID is a catalogue: the discovery, or
// a fixed set of folders.
func TestCatalogue(t *testing.T) {
	t.Parallel()
	var c workflow.Catalogue = catalogue{"infra": {ID: "infra"}}
	assert.Equal(t, "infra", c.Find("infra").ID)
	assert.True(t, c.Find("gone") == nil)
	var _ workflow.Catalogue = &folder.Discovery{}
}

// The run and done events carry the run as the browser reads it.
func TestRunEvent(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(workflow.RunEvent{RunID: "wr1", WorkflowID: "w1", Name: "release", State: "ok", UserName: "ops"})
	assert.NoError(t, err)
	assert.Equal(t, `{"run_id":"wr1","workflow_id":"w1","name":"release","state":"ok","user_name":"ops"}`, string(data))
}
