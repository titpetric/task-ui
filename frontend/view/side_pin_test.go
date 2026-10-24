package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A pinned job reads its args from the pin's JSON list; the run button
// carries them only when there are any.
func TestSidePin(t *testing.T) {
	t.Parallel()
	s := view.SidePin{Pin: model.Pin{Job: "build", Dir: "task", Args: `["a b","c"]`}}
	assert.Equal(t, []string{"a b", "c"}, s.Args())
	assert.Equal(t, `["a b","c"]`, s.ArgsJSON())
	assert.Equal(t, "task/ build a b c", s.Label())

	bare := view.SidePin{Pin: model.Pin{Job: "build", Args: "[]"}}
	assert.Len(t, bare.Args(), 0)
	assert.Equal(t, "", bare.ArgsJSON())
	assert.Equal(t, "build", bare.Label())
	assert.Len(t, view.SidePin{Pin: model.Pin{Args: "nope"}}.Args(), 0)
}
