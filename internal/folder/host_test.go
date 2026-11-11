package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @host:<host> is the root's alone: a job a host, or * for every host,
// that takes the command as its args, with no command of its own and
// no allow list. The bare @host is no job.
func TestFolder_HostJob(t *testing.T) {
	t.Parallel()
	root := &folder.Folder{ID: model.Default, Dir: t.TempDir()}
	job := root.Job("@host:dev1")
	assert.NotNil(t, job)
	assert.True(t, job.IsHost())
	assert.Equal(t, "dev1", job.HostTarget())
	assert.False(t, job.IsShell())
	assert.True(t, job.TakesArgs)
	assert.False(t, job.Interactive)
	assert.Len(t, job.Command, 0)
	assert.Len(t, job.Allow, 0)
	assert.Equal(t, "internal/@host:dev1", job.Key())
	assert.Equal(t, "*", root.Job(model.HostAll).HostTarget())

	assert.Nil(t, root.Job(model.HostBare))
	assert.Nil(t, root.Job(model.HostPrefix))
	assert.Nil(t, root.Job("@host:a b"))
	assert.Nil(t, checkout(t).Job("@host:dev1"))
	assert.Nil(t, (*folder.Folder)(nil).HostJob("dev1"))
	assert.False(t, (*folder.Job)(nil).IsHost())
	assert.Equal(t, "", root.Job("@shell").HostTarget())
	assert.NotNil(t, folder.ErrHostBare)
}
