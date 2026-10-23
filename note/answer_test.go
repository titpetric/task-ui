package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A run's reply is the last one naming it that ended: not the question
// at work, nor a reply of another run.
func TestAnswer(t *testing.T) {
	t.Parallel()
	replies := []model.Note{
		{ID: "a1", RunID: "c1", Body: "first answer"},
		{ID: "q2", RunID: "c2", Body: "and the logs?", State: model.NoteWorking},
	}
	assert.Equal(t, "a1", note.Answer(replies, "c1").ID)
	assert.Nil(t, note.Answer(replies, "c2"), "still at work")
	assert.Nil(t, note.Answer(replies, ""))
	replies = append(replies, model.Note{ID: "a2", RunID: "c2", Body: "the logs say so"})
	assert.Equal(t, "a2", note.Answer(replies, "c2").ID)
}
