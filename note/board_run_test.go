package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// AskRun takes prompt/edit and a run of a folder the user sees, none of
// Claude's own; a board without runs, or a run not there, takes no
// prompt about it.
func TestBoard_AskRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	run := &model.Run{Folder: "agent", Job: "run", UserName: "ops", State: string(model.StateFailed)}
	assert.NoError(t, f.board.Runs.Create(ctx, run))
	internal := &model.Run{Folder: model.Internal, Job: "claude-prompt", UserName: "ops", State: string(model.StateOK)}
	assert.NoError(t, f.board.Runs.Create(ctx, internal))
	claude := &model.Run{Folder: "agent", Job: model.ClaudeJob, UserName: "ops", State: string(model.StateOK)}
	assert.NoError(t, f.board.Runs.Create(ctx, claude))

	_, err := f.board.AskRun(ctx, nil, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = f.board.AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrForbidden)

	assert.NoError(t, f.board.Users.Policies().Put(ctx, "user", f.ops.Name, "prompt/edit", "allow"))
	_, err = f.board.AskRun(ctx, f.ops, internal.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.AskRun(ctx, f.ops, claude.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.AskRun(ctx, f.ops, "nothing", "why")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	// The board has no root, which the place of a prompt is under.
	_, err = f.board.AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = (&note.Board{}).AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
}
