package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// @claude-prompt is every folder's, task-ui agent prompt where the
// folder, or the directory Sub returns, is; no ID resolves to it.
func TestFolder_PromptJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)

	job := f.PromptJob()
	assert.True(t, job.IsPrompt())
	assert.Equal(t, "prompt", job.Kind)
	assert.Equal(t, []string{"agent", "prompt"}, job.Command[1:])
	assert.Equal(t, f.ID+"/@claude-prompt", job.Key())
	assert.Nil(t, f.Job(folder.PromptID), "a prompt starts from the prompt routes")

	job = f.Sub("bin").PromptJob()
	assert.Equal(t, "bin", job.Dir)
	assert.Equal(t, f.ID+"/bin/@claude-prompt", job.Key())
	assert.Contains(t, job.Description, f.ID+"/bin/")

	assert.False(t, f.ShellJob().IsPrompt())
	assert.False(t, (*folder.Job)(nil).IsPrompt())
	assert.True(t, (*folder.Folder)(nil).PromptJob() == nil)
}
