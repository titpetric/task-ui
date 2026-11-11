package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// The question of an answer is the reply before it, or the thread.
func TestQuestion(t *testing.T) {
	t.Parallel()
	thread := &model.Note{ID: "t", Body: "why?"}
	replies := []model.Note{{ID: "a1"}, {ID: "q2", Body: "and then?"}, {ID: "a2"}}
	assert.Equal(t, "why?", note.Question(thread, replies, &replies[0]).Body)
	assert.Equal(t, "and then?", note.Question(thread, replies, &replies[2]).Body)
	assert.Equal(t, "why?", note.Question(thread, replies, &model.Note{ID: "gone"}).Body)
}
