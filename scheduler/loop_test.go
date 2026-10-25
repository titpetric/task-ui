package scheduler_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/workflow"
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

// The workflow runner starts the schedules of workflows.
func TestStarter(t *testing.T) {
	t.Parallel()
	var flows scheduler.Starter = (*workflow.Runner)(nil)
	assert.NotNil(t, &flows)
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	assert.Nil(t, loop.Flows)
	assert.Nil(t, loop.Workflows)
}
