package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestThread_Initials(t *testing.T) {
	t.Parallel()
	thread := view.Thread{Note: model.Note{UserName: "claude"}}
	assert.Equal(t, (&model.User{Name: "claude"}).Initials(), thread.Initials())
}

// A reply of any agent the thread names is that thread's answer; a reply
// to a plain note is a message and no answer.
func TestThread_Answer(t *testing.T) {
	t.Parallel()
	thread := view.Thread{Note: model.Note{ID: "n1", UserName: "ops", State: model.NoteDone}, Answerers: []string{"claude", "consuela"}}
	answer := model.Note{ID: "r1", UserName: "claude"}
	second := model.Note{ID: "r3", UserName: "consuela"}
	assert.True(t, thread.Answer(answer))
	assert.True(t, thread.Answer(second), "the second agent answers too")
	assert.False(t, thread.Answer(model.Note{ID: "r2", UserName: "ops"}))
	plain := view.Thread{Note: model.Note{ID: "n2", UserName: "ops"}, Answerers: []string{"claude", "consuela"}}
	assert.False(t, plain.Answer(answer), "a reply to a plain note is a message")
	assert.False(t, plain.Answer(second))
}
