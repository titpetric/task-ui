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

// Its author or an admin edits a note of a thread; nobody edits the
// agent's answer to an instruction.
func TestThread_Edits(t *testing.T) {
	t.Parallel()
	thread := view.Thread{Note: model.Note{ID: "n1", UserName: "ops", State: model.NoteDone}, Answerer: "claude"}
	answer := model.Note{ID: "r1", UserName: "claude"}
	reply := model.Note{ID: "r2", UserName: "ops"}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	claude := &model.User{Name: "claude", Role: model.RoleRunner}
	assert.True(t, thread.Edits(ops, thread.Note))
	assert.True(t, thread.Edits(ops, reply))
	assert.True(t, thread.Edits(admin, thread.Note))
	assert.False(t, thread.Edits(claude, thread.Note))
	assert.False(t, thread.Edits(claude, answer))
	assert.False(t, thread.Edits(admin, answer))
	assert.False(t, thread.Edits(nil, thread.Note))
	assert.True(t, thread.Answer(answer))
	plain := view.Thread{Note: model.Note{ID: "n2", UserName: "ops"}, Answerer: "claude"}
	assert.False(t, plain.Answer(answer), "a reply to a plain note is a message")
	assert.True(t, plain.Edits(claude, answer))
}
