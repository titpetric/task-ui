package note_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
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
	own := &model.Run{Folder: "agent", Job: model.ClaudeJob, UserName: "ops", State: string(model.StateOK)}
	assert.NoError(t, f.board.Runs.Create(ctx, own))

	_, err := f.board.AskRun(ctx, nil, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = f.board.AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrForbidden)

	assert.NoError(t, f.board.Users.Policies().Put(ctx, "user", f.ops.Name, "prompt/edit", "allow"))
	_, err = f.board.AskRun(ctx, f.ops, internal.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.AskRun(ctx, f.ops, own.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.AskRun(ctx, f.ops, "nothing", "why")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	// The board has no root, which the place of a prompt is under.
	_, err = f.board.AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = (&note.Board{}).AskRun(ctx, f.ops, run.ID, "why")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
}

// A prompt about a run whose recording expired reads that the run has
// no output any more.
func TestBoard_AskRunExpired(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `cat "$TASK_UI_PROMPT_RUN" > "$TASK_UI_PROMPT_REPLY"`}
	run := &model.Run{Folder: "proj", Job: "build", UserName: "admin", State: string(model.StateOK), CastPath: "proj/build/gone.cast"}
	assert.NoError(t, f.board.Runs.Create(ctx, run))

	asked, err := f.board.AskRun(ctx, admin, run.ID, "what did it print")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteDone, f.settled(t, asked.ID).State)
	assert.Contains(t, replies(t, f, asked.ID)[0].Body, recording.Expired)
}

// SummarizeBatch asks about a batch's first run with the hosts'
// question, in a thread about that run.
func TestBoard_SummarizeBatch(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	run := &model.Run{Folder: model.Default, Job: model.HostJob("dev1"), Target: "dev1", UserName: "admin", State: string(model.StateOK)}
	assert.NoError(t, f.board.Runs.Create(ctx, run))

	asked, err := f.board.SummarizeBatch(ctx, admin, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, note.BatchSummaryQuestion, asked.Body)
	assert.Equal(t, run.ID, asked.AboutRunID)
	f.settled(t, asked.ID)
}

// SummarizeRun asks about a run with the short question; the prompt's
// run gets the summary's shape in TASK_UI_PROMPT_INSTRUCTIONS, and a
// reply in the thread asks without it.
func TestBoard_SummarizeRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `printf '%s|%s' "$TASK_UI_PROMPT_INSTRUCTIONS" "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}
	run := &model.Run{Folder: "proj", Job: "build", UserName: "admin", State: string(model.StateOK)}
	assert.NoError(t, f.board.Runs.Create(ctx, run))

	asked, err := f.board.SummarizeRun(ctx, admin, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, note.SummaryQuestion, asked.Body)
	assert.Equal(t, run.ID, asked.AboutRunID)
	assert.Equal(t, model.NoteDone, f.settled(t, asked.ID).State)
	assert.Equal(t, claude.RunSummary+"|"+note.SummaryQuestion, replies(t, f, asked.ID)[0].Body)

	_, err = f.board.Ask(ctx, admin, asked.ID, "", "and the images")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool { return len(replies(t, f, asked.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	assert.True(t, strings.HasPrefix(replies(t, f, asked.ID)[2].Body, "|admin:\n"+note.SummaryQuestion))
}
