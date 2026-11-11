package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// The column is what a user typed, so the spaces around a marker go and
// an empty entry is not a marker that matches everything.
func TestMarkers(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"Next:", "TODO:"}, model.Markers(" Next: , TODO: ,, "))
	assert.Equal(t, []string(nil), model.Markers(" , "))
}

// A marker matches anywhere in the step's output, case ignored, because
// Claude writes "next:" as readily as "Next:". The marker that answers
// is the one the user wrote, for the check's answer to quote.
func TestMarked(t *testing.T) {
	t.Parallel()
	markers := model.Markers("Next:,TODO:")
	assert.Equal(t, "Next:", model.Marked(markers, "done.\nnext: the second half\n"))
	assert.Equal(t, "TODO:", model.Marked(markers, "TODO: write it up"))
	assert.Equal(t, "", model.Marked(markers, "all done, nothing left"))
	assert.Equal(t, "", model.Marked(nil, "Next: anything"))
}

func TestPasses(t *testing.T) {
	t.Parallel()
	assert.Equal(t, model.CheckPasses, model.Passes(0))
	assert.Equal(t, model.LeastPasses, model.Passes(1))
	assert.Equal(t, 5, model.Passes(5))
	assert.Equal(t, model.MostPasses, model.Passes(500))
}
