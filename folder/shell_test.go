package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @shell is every folder's: bash -l with a terminal where the folder,
// or the directory Sub returns, is, for admins only.
func TestFolder_ShellJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	f.Allow = []string{"ops"}

	job := f.Job(folder.ShellID)
	assert.NotNil(t, job)
	assert.True(t, job.IsShell())
	assert.True(t, job.Interactive)
	assert.Equal(t, "shell", job.Kind)
	assert.Equal(t, []string{"bash", "-l"}, job.Command)
	assert.Equal(t, []string{model.RoleAdmin}, job.Allow)
	assert.Equal(t, "", job.Dir)
	assert.Equal(t, f.ID+"/@shell", job.Key())

	sub := f.Sub("bin")
	assert.NotNil(t, sub)
	job = sub.Job(folder.ShellID)
	assert.Equal(t, "bin", job.Dir)
	assert.Equal(t, f.ID+"/bin/@shell", job.Key())
	assert.Contains(t, job.Description, f.ID+"/bin/")

	assert.False(t, f.Job("@git:status").IsShell())
	assert.False(t, (*folder.Job)(nil).IsShell())
	assert.True(t, (*folder.Folder)(nil).ShellJob() == nil)
	assert.Contains(t, folder.ErrShellUnattended.Error(), "no schedule or workflow step")
}
