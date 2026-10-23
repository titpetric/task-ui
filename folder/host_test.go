package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @host is the root's alone: a job that takes the command as its args,
// with no command of its own and no allow list.
func TestFolder_HostJob(t *testing.T) {
	t.Parallel()
	root := &folder.Folder{ID: model.Default, Dir: t.TempDir()}
	job := root.Job(folder.HostID)
	assert.NotNil(t, job)
	assert.True(t, job.IsHost())
	assert.False(t, job.IsShell())
	assert.True(t, job.TakesArgs)
	assert.False(t, job.Interactive)
	assert.Len(t, job.Command, 0)
	assert.Len(t, job.Allow, 0)
	assert.Equal(t, "_/@host", job.Key())

	assert.Nil(t, checkout(t).Job(folder.HostID))
	assert.Nil(t, (*folder.Folder)(nil).HostJob())
	assert.False(t, (*folder.Job)(nil).IsHost())
}
