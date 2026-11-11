package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @forgejo-sync is the root's alone: task-ui forgejo-sync, for admins,
// with --dry-run as its argument.
func TestFolder_ForgejoSyncJob(t *testing.T) {
	t.Parallel()
	root := &folder.Folder{ID: model.Default, Dir: t.TempDir()}
	job := root.Job(folder.ForgejoSyncID)
	assert.NotNil(t, job)
	assert.False(t, job.IsShell())
	assert.False(t, job.Interactive)
	assert.True(t, job.TakesArgs)
	assert.Equal(t, "forgejo-sync", job.Command[len(job.Command)-1])
	assert.Equal(t, []string{model.RoleAdmin}, job.Allow)
	assert.Equal(t, "internal/@forgejo-sync", job.Key())

	assert.Nil(t, checkout(t).Job(folder.ForgejoSyncID))
	assert.Nil(t, (*folder.Folder)(nil).ForgejoSyncJob())
}
