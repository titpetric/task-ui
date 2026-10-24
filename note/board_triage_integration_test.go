//go:build integration

package note_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// TriageRun opens the triage thread of a failed run on the run's page,
// with the run and the triage instruction, waits for Claude's answer,
// reads its verdict and leaves the note on the board as the schedule's
// user, with the run's link.
func TestBoard_TriageRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.BaseURL = "https://task.example"
	f.board.AgentCommand = []string{"sh", "-c", `{ printf 'Q: %s\n' "$TASK_UI_PROMPT"; printf 'I: %s\n' "$TASK_UI_PROMPT_INSTRUCTIONS"; head -n 2 "$TASK_UI_PROMPT_RUN"; printf '\n**Result: resolved**\nAdded the file.\n'; } > "$TASK_UI_PROMPT_REPLY"`}

	failed := ran(t, f, admin, "deploy", "echo no config; exit 2")
	assert.Equal(t, model.StateFailed, failed.State)

	posted, verdict, err := f.board.TriageRun(ctx, admin, failed.ID)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict)
	assert.Equal(t, "Attempted triage of failing proj/deploy task, result: resolved. Info: [https://task.example/runs/"+failed.ID+"](https://task.example/runs/"+failed.ID+")", posted.Body)
	assert.Equal(t, admin.Name, posted.UserName)
	assert.False(t, posted.IsPrompt(), "a note of the board")

	threads, err := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutRunID: failed.ID})
	assert.NoError(t, err)
	assert.Len(t, threads, 1)
	thread := threads[0]
	assert.Equal(t, note.TriageQuestion, thread.Body)
	assert.Equal(t, "proj", thread.Dir)
	assert.Equal(t, model.NoteDone, thread.State)
	answer := replies(t, f, thread.ID)
	assert.Len(t, answer, 1)
	assert.True(t, strings.HasPrefix(answer[0].Body, "Q: "+note.TriageQuestion+"\nI: This is the triage of a scheduled run that failed."), answer[0].Body)
	assert.Contains(t, answer[0].Body, "\nRun "+failed.ID+" of the job deploy in proj, by admin.\n")

	board, err := f.board.Notes.List(ctx, storage.NoteFilter{})
	assert.NoError(t, err)
	assert.Len(t, board, 1, "the note, and no thread, on the board")
}

// A triage whose prompt fails is not resolved, and noted so; a run no
// prompt reaches, or a user without prompt/edit, ends the triage before
// any note.
func TestBoard_TriageRun_refused(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `echo "Result: resolved" > "$TASK_UI_PROMPT_REPLY"; exit 3`}

	failed := ran(t, f, admin, "deploy", "exit 2")
	posted, verdict, err := f.board.TriageRun(ctx, admin, failed.ID)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageUnresolved, verdict)
	assert.Contains(t, posted.Body, "result: not resolved. Info: [/runs/"+failed.ID+"]")

	_, _, err = f.board.TriageRun(ctx, f.ops, failed.ID)
	assert.ErrorIs(t, err, note.ErrForbidden)
	_, _, err = f.board.TriageRun(ctx, admin, model.NewID())
	assert.ErrorIs(t, err, storage.ErrNotFound)
	board, err := f.board.Notes.List(ctx, storage.NoteFilter{})
	assert.NoError(t, err)
	assert.Len(t, board, 1, "the refused triages leave no note")
}
