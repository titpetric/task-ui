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
