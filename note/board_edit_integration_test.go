//go:build integration

package note_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/runner"
)

// answered waits for the only reply of a thread to say want, and for the
// thread to settle, and returns the reply.
func answered(t *testing.T, f *fixture, id, want string) model.Note {
	t.Helper()
	var last []model.Note
	assert.Eventually(t, func() bool {
		last = replies(t, f, id)
		root, err := f.board.Notes.Get(context.Background(), id)
		return err == nil && root.Finished() && len(last) == 1 && last[0].Body == want && last[0].State == ""
	}, 10*time.Second, 20*time.Millisecond)
	if len(last) != 1 {
		t.Fatalf("replies %+v", last)
	}
	return last[0]
}

// Editing an instruction runs it again: the agent's answer takes the new
// one in place, from a run under the same ID, and the thread gets no new
// reply.
func TestBoard_EditInstruction(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	posted, err := f.board.Post(ctx, f.ops, "", "@claude answer me")
	assert.NoError(t, err)
	first := answered(t, f, posted.ID, "Done: @claude answer me")
	assert.NotEqual(t, "", first.RunID)

	edited, err := f.board.Edit(ctx, f.ops, posted.ID, "@claude answer me, in French")
	assert.NoError(t, err)
	assert.Equal(t, "@claude answer me, in French", edited.Body)
	assert.NotNil(t, edited.EditedAt)
	again := answered(t, f, posted.ID, "Done: @claude answer me, in French")
	assert.Equal(t, first.ID, again.ID, "the same reply")
	assert.Equal(t, first.RunID, again.RunID, "the same run")
	root, _ := f.board.Notes.Get(ctx, posted.ID)
	assert.Equal(t, model.NoteDone, root.State)
	assert.Equal(t, first.RunID, root.RunID)
	session, ok := f.board.Runner.Get(first.RunID)
	assert.True(t, ok)
	run := session.Snapshot()
	cast, err := f.board.Recordings.Read(&run)
	assert.NoError(t, err)
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	assert.NoError(t, err)
	assert.Contains(t, replay.Output.Text(), "in French")
}

// Editing a plain note runs nothing; only its author or an admin edits,
// and the agent's answer changes only by its question.
func TestBoard_EditMessage(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	plain, err := f.board.Post(ctx, f.ops, "", "have a look")
	assert.NoError(t, err)
	edited, err := f.board.Edit(ctx, f.ops, plain.ID, "have a look, @alice")
	assert.NoError(t, err)
	assert.Equal(t, "", edited.State)
	assert.True(t, edited.Mentioned("alice"))
	assert.NotNil(t, edited.EditedAt)
	time.Sleep(50 * time.Millisecond)
	assert.Len(t, replies(t, f, plain.ID), 0)

	other := &model.User{Name: "alice", Role: model.RoleRunner}
	assert.NoError(t, f.board.Users.Create(ctx, other))
	_, err = f.board.Edit(ctx, other, plain.ID, "mine now")
	assert.ErrorIs(t, err, note.ErrNotAuthor)
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, f.board.Users.Create(ctx, admin))
	_, err = f.board.Edit(ctx, admin, plain.ID, "tidied by an admin")
	assert.NoError(t, err)
	_, err = f.board.Edit(ctx, f.ops, plain.ID, " ")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = f.board.Edit(ctx, f.ops, "missing", "x")
	assert.Error(t, err)

	asked, err := f.board.Post(ctx, f.ops, "", "@claude answer this")
	assert.NoError(t, err)
	answer := answered(t, f, asked.ID, "Done: @claude answer this")
	claude, _ := f.board.Users.GetByName(ctx, "claude")
	_, err = f.board.Edit(ctx, claude, answer.ID, "something else")
	assert.ErrorIs(t, err, note.ErrInvalid)
}

// An edit during a rerun waits for that run, then runs again on the
// latest text, in the same reply and run.
func TestBoard_EditWhileRunning(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	posted, err := f.board.Post(ctx, f.ops, "", "@claude answer")
	assert.NoError(t, err)
	first := answered(t, f, posted.ID, "Done: @claude answer")

	_, err = f.board.Edit(ctx, f.ops, posted.ID, "@claude answer slow")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool {
		got := replies(t, f, posted.ID)
		return len(got) == 1 && got[0].State == model.NoteWorking
	}, 10*time.Second, 20*time.Millisecond)
	_, err = f.board.Edit(ctx, f.ops, posted.ID, "@claude answer twice")
	assert.NoError(t, err)
	last := answered(t, f, posted.ID, "Done: @claude answer twice")
	assert.Equal(t, first.ID, last.ID)
	assert.Equal(t, first.RunID, last.RunID)
}

// Editing a prompt, or a reply in its thread, asks again: Claude's reply
// to it changes in place, from the thread up to it, and the replies
// before it stay. An answer whose run is of another job runs again under
// a new run ID.
func TestBoard_EditPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	notes := f.board.AgentCommand[2]
	admin := f.withPrompts(t)
	// One command answers both: the Notes agent's as the fixture's, a
	// prompt's as withPrompts has it.
	f.board.AgentCommand[2] = `if [ "$0" = note ]; then ` + notes + `; else ` + f.board.AgentCommand[2] + `; fi`
	ctx := context.Background()
	asked, err := f.board.Ask(ctx, admin, "", "proj", "what builds here")
	assert.NoError(t, err)
	prefix := "In " + f.board.Catalogue.Find("proj").Dir + ": "
	first := answered(t, f, asked.ID, prefix+"what builds here")

	_, err = f.board.Edit(ctx, f.ops, asked.ID, "mine")
	assert.ErrorIs(t, err, note.ErrNotAuthor)
	_, err = f.board.Edit(ctx, admin, asked.ID, "what runs here")
	assert.NoError(t, err)
	again := answered(t, f, asked.ID, prefix+"what runs here")
	assert.Equal(t, first.ID, again.ID)
	assert.Equal(t, first.RunID, again.RunID)

	reply, err := f.board.Post(ctx, admin, asked.ID, "and which first")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool { return len(replies(t, f, asked.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, model.NoteDone, f.settled(t, asked.ID).State)
	before := replies(t, f, asked.ID)
	_, err = f.board.Edit(ctx, admin, reply.ID, "and which last")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool {
		got := replies(t, f, asked.ID)
		return len(got) == 3 && got[2].State == "" && got[2].Body != before[2].Body
	}, 10*time.Second, 20*time.Millisecond)
	after := replies(t, f, asked.ID)
	assert.Equal(t, before[2].ID, after[2].ID)
	assert.Equal(t, before[2].RunID, after[2].RunID)
	assert.Equal(t, before[0].Body, after[0].Body, "the answer before it stays")
	assert.Contains(t, after[2].Body, "admin:\nand which last")
	assert.NotContains(t, after[2].Body, "and which first")

	posted, err := f.board.Post(ctx, f.ops, "", "@claude answer")
	assert.NoError(t, err)
	answer := answered(t, f, posted.ID, "Done: @claude answer")
	assert.NoError(t, f.board.Notes.Answer(ctx, answer.ID, answer.Body, "", first.RunID))
	_, err = f.board.Edit(ctx, f.ops, posted.ID, "@claude answer anew")
	assert.NoError(t, err)
	moved := answered(t, f, posted.ID, "Done: @claude answer anew")
	assert.Equal(t, answer.ID, moved.ID)
	assert.NotEqual(t, first.RunID, moved.RunID, "another job's run is no run to reuse")
	kept, ok := f.board.Runner.Get(first.RunID)
	assert.True(t, ok)
	assert.Equal(t, "proj/"+model.PromptJob, snapshotKey(kept), "the prompt's run stays")
}

// Claude's answers to prompts carry claude; an instruction's carry the
// agent its thread names, and a thread naming none has no answerer.
func TestBoard_Answerer(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "claude", (&note.Board{}).Answerer())
	f := newFixture(t, "agent/nothing")
	ctx := context.Background()
	assert.Equal(t, "claude", f.board.AnswererOf(ctx, &model.Note{Body: "@claude go", UserName: "ops"}))
	assert.Equal(t, "", f.board.AnswererOf(ctx, &model.Note{Body: "@ops go", UserName: "claude"}), "ops runs no agent")
	assert.Equal(t, "", f.board.AnswererOf(ctx, &model.Note{Body: "@claude go", UserName: "claude"}), "its own note")
}
