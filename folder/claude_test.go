package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @claude is every folder's: task-ui agent shell with a terminal where
// the folder, or the directory Sub returns, is; a shell, for admins only.
func TestFolder_ClaudeJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	f.Allow = []string{"ops"}

	job := f.Job(folder.ClaudeID)
	assert.NotNil(t, job)
	assert.True(t, job.IsShell(), "the shell permission gates it, as @shell")
	assert.True(t, job.Interactive)
	assert.Equal(t, "claude", job.Name)
	assert.Equal(t, []string{"agent", "shell"}, job.Command[1:])
	assert.Equal(t, []string{model.RoleAdmin}, job.Allow)
	assert.Equal(t, f.ID+"/@claude", job.Key())

	job = f.Sub("bin").Job(folder.ClaudeID)
	assert.Equal(t, "bin", job.Dir)
	assert.Equal(t, f.ID+"/bin/@claude", job.Key())
	assert.Contains(t, job.Description, f.ID+"/bin/")

	assert.False(t, job.IsPrompt())
	assert.True(t, (*folder.Folder)(nil).ClaudeJob() == nil)
}
