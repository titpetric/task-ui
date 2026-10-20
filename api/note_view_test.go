package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A note's view carries the note's fields and its replies.
func TestNoteView(t *testing.T) {
	t.Parallel()
	view := api.NoteView{Note: model.Note{ID: "n1", Body: "hi"}, Replies: []model.Note{{ID: "r1", Body: "hello"}}}
	data, err := json.Marshal(view)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"id":"n1"`)
	assert.Contains(t, string(data), `"replies":[{"id":"r1"`)
}
