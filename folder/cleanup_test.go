package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @cleanup is the root's alone: task-ui cleanup, for admins, with no one
// at the keyboard.
func TestFolder_CleanupJob(t *testing.T) {
	t.Parallel()
	root := &folder.Folder{ID: model.Default, Dir: t.TempDir()}
	job := root.Job(folder.CleanupID)
	assert.NotNil(t, job)
	assert.False(t, job.IsShell())
	assert.False(t, job.IsHost())
	assert.False(t, job.Interactive)
	assert.True(t, job.TakesArgs)
	assert.Equal(t, "cleanup", job.Command[len(job.Command)-1])
	assert.Equal(t, []string{model.RoleAdmin}, job.Allow)
	assert.Equal(t, "internal/@cleanup", job.Key())

	assert.Nil(t, checkout(t).Job(folder.CleanupID))
	assert.Nil(t, (*folder.Folder)(nil).CleanupJob())
}
