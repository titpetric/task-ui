package note

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The run of a thread sits under the run already working it, so a thread
// of several asks reads as one run; a thread a workflow run's step or a
// run's summary asked sits under that run, and a prompt somebody typed
// sits under nothing.
func TestParentOf(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		thread *model.Note
		id     string
		kind   string
	}{
		{"none", nil, "", ""},
		{"asked", &model.Note{ID: "n1", Body: "why"}, "", ""},
		{"again", &model.Note{ID: "n1", RunID: "r1"}, "r1", storage.ParentRun},
		{"step", &model.Note{ID: "n1", AboutWorkflowRunID: "wr1"}, "wr1", storage.ParentWorkflowRun},
		{"summary", &model.Note{ID: "n1", AboutRunID: "r2"}, "r2", storage.ParentRun},
		{"working", &model.Note{ID: "n1", RunID: "r1", AboutRunID: "r2"}, "r1", storage.ParentRun},
	} {
		id, kind := parentOf(test.thread)
		assert.Equal(t, test.id, id, test.name)
		assert.Equal(t, test.kind, kind, test.name)
	}
}
