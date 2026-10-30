package phpscript_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestDispatcher(t *testing.T) {
	t.Parallel()
	var dispatcher phpscript.Dispatcher = &fakeDispatcher{id: "01JRUN"}
	m := &phpscript.Module{Vendor: "titpetric", Plugin: "refinement", Root: "/x"}

	id, err := dispatcher.Dispatch(t.Context(), m, "use ULID as primary keys", "action/refinement.md", "01THREAD")
	assert.NoError(t, err)
	assert.Equal(t, "01JRUN", id)
	assert.Equal(t, []string{"titpetric/refinement:action/refinement.md:use ULID as primary keys"},
		dispatcher.(*fakeDispatcher).calls)
	assert.Equal(t, []string{"01THREAD"}, dispatcher.(*fakeDispatcher).folders,
		"the folder is the thread's scratch space in the artifact tree")
}
