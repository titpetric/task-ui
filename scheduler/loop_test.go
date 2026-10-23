package scheduler_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/scheduler"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func TestNewLoop(t *testing.T) {
	t.Parallel()
	var folders scheduler.Catalogue = catalogue{}
	loop := scheduler.NewLoop(nil, nil, folders, nil)
	assert.Equal(t, time.Minute, loop.Interval)
	assert.NotNil(t, loop.Now)
}
