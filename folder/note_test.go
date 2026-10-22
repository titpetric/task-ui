package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// @claude-note is the Notes agent, task-ui agent note where the folder
// is; the board runs it as the root's, and no ID resolves to it.
func TestFolder_NoteJob(t *testing.T) {
	t.Parallel()
	root := &folder.Folder{ID: model.Default, Title: model.Default, Dir: t.TempDir()}

	job := root.NoteJob()
	assert.True(t, job.IsNote())
	assert.False(t, job.IsShell())
	assert.Equal(t, "note", job.Kind)
	assert.Equal(t, []string{"agent", "note"}, job.Command[1:])
	assert.Equal(t, "_/@claude-note", job.Key())
	assert.Nil(t, root.Job(folder.NoteID), "the board starts it, on a note")
	assert.Nil(t, checkout(t).Job(folder.NoteID))

	assert.False(t, root.PromptJob().IsNote())
	assert.False(t, (*folder.Job)(nil).IsNote())
	assert.True(t, (*folder.Folder)(nil).NoteJob() == nil)
}
