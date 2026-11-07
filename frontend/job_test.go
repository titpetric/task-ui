package frontend

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A job page reads the summaries of the rows it lists, a run's own page
// the summary of the single run it shows and of no rows: without that
// run the summary under way is read nowhere and Summarize shows again.
func TestRunIDs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, len(runIDs(&view.Page{})))

	rows := &view.Page{Runs: []view.Run{{Run: model.Run{ID: "r1"}}, {Run: model.Run{ID: "r2"}}}}
	assert.Equal(t, []string{"r1", "r2"}, runIDs(rows))

	one := &view.Page{Run: &view.Run{Run: model.Run{ID: "r3"}}}
	assert.Equal(t, []string{"r3"}, runIDs(one))
}
