package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @script is every folder's, for a workflow's script step: it takes the
// command as its args, runs where the folder or its directory is, and
// stays out of Job, so no run outside a workflow starts it.
func TestFolder_ScriptJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)

	job := f.ScriptJob()
	assert.True(t, job.IsScript())
	assert.True(t, job.TakesArgs)
	assert.Equal(t, "script", job.Kind)
	assert.Equal(t, []string{model.RoleAdmin}, job.Allow)
	assert.Equal(t, f.ID+"/@script", job.Key())
	assert.Nil(t, f.Job(folder.ScriptID))

	sub := f.Sub("bin")
	assert.Equal(t, "bin", sub.ScriptJob().Dir)
	assert.Contains(t, sub.ScriptJob().Description, f.ID+"/bin/")

	assert.False(t, f.ShellJob().IsScript())
	assert.False(t, (*folder.Job)(nil).IsScript())
	assert.True(t, (*folder.Folder)(nil).ScriptJob() == nil)
	assert.Contains(t, folder.ErrScriptEmpty.Error(), "command")
}
