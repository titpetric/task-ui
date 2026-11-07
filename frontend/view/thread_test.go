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

// Its author or an admin edits a note of a thread; nobody edits an
// agent's answer to an instruction, whichever of the agents wrote it.
func TestThread_Edits(t *testing.T) {
	t.Parallel()
	thread := view.Thread{Note: model.Note{ID: "n1", UserName: "ops", State: model.NoteDone}, Answerers: []string{"claude", "consuela"}}
	answer := model.Note{ID: "r1", UserName: "claude"}
	second := model.Note{ID: "r3", UserName: "consuela"}
	reply := model.Note{ID: "r2", UserName: "ops"}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	claude := &model.User{Name: "claude", Role: model.RoleRunner}
	consuela := &model.User{Name: "consuela", Role: model.RoleRunner}
	assert.True(t, thread.Edits(ops, thread.Note))
	assert.True(t, thread.Edits(ops, reply))
	assert.True(t, thread.Edits(admin, thread.Note))
	assert.False(t, thread.Edits(claude, thread.Note))
	assert.False(t, thread.Edits(claude, answer))
	assert.False(t, thread.Edits(admin, answer))
	assert.False(t, thread.Edits(nil, thread.Note))
	assert.True(t, thread.Answer(answer))
	assert.True(t, thread.Answer(second), "the second agent answers too")
	assert.False(t, thread.Edits(consuela, second))
	assert.False(t, thread.Edits(admin, second))
	plain := view.Thread{Note: model.Note{ID: "n2", UserName: "ops"}, Answerers: []string{"claude", "consuela"}}
	assert.False(t, plain.Answer(answer), "a reply to a plain note is a message")
	assert.False(t, plain.Answer(second))
	assert.True(t, plain.Edits(claude, answer))
}
