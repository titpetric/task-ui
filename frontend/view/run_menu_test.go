package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func menuJob(id, desc string, steps ...string) *folder.Job {
	return &folder.Job{ID: id, Target: detector.Target{Name: id, Description: desc, Steps: steps}}
}

// A menu lists default first and the rest by ID, case aside; the button
// runs the default, the only job, or the job given.
func TestRunMenu(t *testing.T) {
	t.Parallel()
	f := &folder.Folder{ID: "web"}
	zap, build, def, lint := menuJob("zap", ""), menuJob("Build", ""), menuJob("default", "Everything"), menuJob("lint", "")

	m := view.NewRunMenu(f, []*folder.Job{zap, build, def, lint}, nil)
	assert.Equal(t, []*folder.Job{def, build, lint, zap}, m.Jobs)
	assert.Equal(t, def, m.Main)
	assert.True(t, m.Menu())

	none := view.NewRunMenu(f, []*folder.Job{zap, lint}, nil)
	assert.True(t, none.Main == nil, "no default: the button opens the menu")
	assert.True(t, none.Menu())

	one := view.NewRunMenu(f, []*folder.Job{lint}, nil)
	assert.Equal(t, lint, one.Main)
	assert.False(t, one.Menu(), "one job runs directly")

	file := menuJob("./tool.sh", "")
	given := view.NewRunMenu(f, []*folder.Job{lint}, file)
	assert.Equal(t, file, given.Main)
	assert.True(t, given.Menu(), "the target is not the file")

	assert.True(t, view.NewRunMenu(f, nil, nil) == nil)
	assert.True(t, view.NewRunMenu(nil, []*folder.Job{lint}, nil) == nil)

	assert.Equal(t, "Everything", m.Describe(def))
	assert.Equal(t, "Run the lint target", m.Describe(lint))
	assert.Equal(t, "Run the push target", m.Describe(menuJob("push", "run: docker push", "run: docker push")))
	assert.Equal(t, "Run the push target", m.Describe(menuJob("push", "push")))
}
