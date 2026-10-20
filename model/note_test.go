package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestMentions(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"claude", "ops"}, model.Mentions("@claude please check, and tell @ops. Thanks @claude."))
	assert.Equal(t, []string{"a.b"}, model.Mentions("(@a.b)"))
	assert.Len(t, model.Mentions("mail me at me@example.com"), 0)
	assert.Len(t, model.Mentions("nothing here"), 0)
}

func TestNote_Mentioned(t *testing.T) {
	t.Parallel()
	note := &model.Note{Mentions: model.EncodeMentions([]string{"claude", "ops"})}
	assert.Equal(t, " claude ops ", note.Mentions)
	assert.True(t, note.Mentioned("claude"))
	assert.False(t, note.Mentioned("clau"))
	assert.False(t, note.Mentioned(""))
	assert.Equal(t, []string{"claude", "ops"}, note.MentionList())
	assert.Equal(t, "", model.EncodeMentions(nil))
}

func TestNote_Finished(t *testing.T) {
	t.Parallel()
	assert.False(t, (&model.Note{State: model.NoteWorking}).Finished())
	assert.True(t, (&model.Note{State: model.NoteDone}).Finished())
	assert.True(t, (&model.Note{State: model.NoteFailed}).Finished())
}
