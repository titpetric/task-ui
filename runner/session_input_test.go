package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestSession_TakeInput(t *testing.T) {
	t.Parallel()
	run := &model.Run{ID: model.NewID(), Folder: "f", Job: "build", TermCols: 80, TermRows: 24}
	session := runner.NewSession(run, &folder.Folder{ID: "f"}, &folder.Job{ID: "build"})

	// A job that takes no input refuses keystrokes until someone takes it.
	assert.ErrorIs(t, session.Input([]byte("yes\n")), runner.ErrNotInteractive)
	assert.Equal(t, "", session.InputBy())

	assert.True(t, session.TakeInput("black"))
	assert.False(t, session.TakeInput("other"), "the first taker is the one the run names")
	assert.Equal(t, "black", session.InputBy())
	assert.NoError(t, session.Input([]byte("yes\n")))
}

func TestSession_Input_interactive(t *testing.T) {
	t.Parallel()
	run := &model.Run{ID: model.NewID(), Folder: "f", Job: "@shell", TermCols: 80, TermRows: 24}
	job := &folder.Job{ID: "@shell"}
	job.Interactive = true
	session := runner.NewSession(run, &folder.Folder{ID: "f"}, job)
	assert.NoError(t, session.Input([]byte("ls\n")))
	assert.Equal(t, "", session.InputBy())
}
