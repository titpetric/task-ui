package note_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	list := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		list = append(list, c[id])
	}
	return list
}

// fixture is a board over a private database, with the users ops and
// claude and an agent job that echoes its note, and writes the reply file
// for a note with the word answer.
type fixture struct {
	board *note.Board
	ops   *model.User
}

func newFixture(t *testing.T, agentJob string) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, ops))
	assert.NoError(t, users.Create(ctx, &model.User{Name: "claude", Role: model.RoleRunner}))

	dir := filepath.Join(t.TempDir(), "agent")
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	f := &folder.Folder{ID: "agent", Title: "agent", Dir: dir}
	job := &folder.Job{Folder: "agent", ID: "run", Kind: "script", Command: []string{"sh", "-c", `echo "note $TASK_UI_NOTE_ID from $TASK_UI_NOTE_AUTHOR: $TASK_UI_NOTE"; case "$TASK_UI_NOTE" in *answer*) echo "Done: $TASK_UI_NOTE" > "$TASK_UI_NOTE_REPLY";; esac; test "$TASK_UI_NOTE" != "@claude fail" && test "$TASK_UI_NOTE" != "@claude answer, fail"`}}
	job.Name = "run"
	f.Jobs = []*folder.Job{job}

	recordings := &recording.Store{Root: t.TempDir()}
	manager := runner.New(runner.Options{Runs: storage.NewRunStorage(db), Recordings: recordings, Grace: time.Second, KillTimeout: 5 * time.Second, Keep: time.Minute})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	board := &note.Board{
		Notes: storage.NewNoteStorage(db), Users: users, Catalogue: catalogue{"agent": f},
		Runner: manager, Recordings: recordings, Events: event.NewBus(), Agent: "claude", AgentJob: agentJob,
	}
	return &fixture{board: board, ops: ops}
}

// settled waits for a note to leave the open and working states.
func (f *fixture) settled(t *testing.T, id string) *model.Note {
	t.Helper()
	var got *model.Note
	assert.Eventually(t, func() bool {
		var err error
		got, err = f.board.Notes.Get(context.Background(), id)
		return err == nil && got.Finished()
	}, 10*time.Second, 20*time.Millisecond)
	return got
}

func replies(t *testing.T, f *fixture, id string) []model.Note {
	t.Helper()
	all, err := f.board.Notes.Replies(context.Background(), []string{id})
	assert.NoError(t, err)
	return all[id]
}

// A note naming the agent runs the agent job with the note in its
// environment; the agent replies with how the run ended and its output.
func TestBoard_Instruction(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()

	posted, err := f.board.Post(ctx, f.ops, "", "@claude update the weather job")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, posted.State)

	done := f.settled(t, posted.ID)
	assert.Equal(t, model.NoteDone, done.State)
	assert.NotEqual(t, "", done.RunID)
	reply := replies(t, f, posted.ID)
	assert.Len(t, reply, 1)
	assert.Equal(t, "claude", reply[0].UserName)
	assert.Contains(t, reply[0].Body, "ended ok")
	assert.Contains(t, reply[0].Body, "note "+posted.ID+" from ops: @claude update the weather job")

	failed, err := f.board.Post(ctx, f.ops, "", "@claude fail")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, f.settled(t, failed.ID).State)
}

// A reply naming the agent opens its note again, and the agent works from
// the whole thread.
func TestBoard_ReplyInstruction(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	first, err := f.board.Post(ctx, f.ops, "", "@claude update the weather job")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteDone, f.settled(t, first.ID).State)

	reply, err := f.board.Post(ctx, f.ops, first.ID, "@claude and add Lisbon")
	assert.NoError(t, err)
	assert.Equal(t, "", reply.State, "the reply carries no state, its note does")
	assert.Eventually(t, func() bool { return len(replies(t, f, first.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	again := f.settled(t, first.ID)
	assert.Equal(t, model.NoteDone, again.State)
	last := replies(t, f, first.ID)[2]
	assert.Equal(t, "claude", last.UserName)
	assert.Contains(t, last.Body, "ops:")
	assert.Contains(t, last.Body, "@claude and add Lisbon")
}

// A note to someone else, a reply that names no agent, a viewer's note and
// the agent's own note are messages, not instructions.
func TestBoard_Messages(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	claude, err := f.board.Users.GetByName(ctx, "claude")
	assert.NoError(t, err)

	other, err := f.board.Post(ctx, f.ops, "", "@alice have a look")
	assert.NoError(t, err)
	assert.Equal(t, "", other.State)
	reply, err := f.board.Post(ctx, f.ops, other.ID, "and you, @alice")
	assert.NoError(t, err)
	assert.Equal(t, "", reply.State)
	own, err := f.board.Post(ctx, claude, "", "@claude a reminder")
	assert.NoError(t, err)
	assert.Equal(t, "", own.State)

	viewer := &model.User{Name: "guest", Role: model.RoleViewer}
	assert.NoError(t, f.board.Users.Create(ctx, viewer))
	watched, err := f.board.Post(ctx, viewer, "", "@claude delete everything")
	assert.NoError(t, err)
	assert.Equal(t, "", watched.State, "a viewer cannot hand out work")

	_, err = f.board.Post(ctx, f.ops, "", "  ")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = f.board.Post(ctx, f.ops, reply.ID, "a reply to a reply")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = f.board.Post(ctx, f.ops, "missing", "x")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = f.board.Post(ctx, nil, "", "x")
	assert.ErrorIs(t, err, note.ErrInvalid)
}

// Without an agent job an instruction stays open, and an agent reports on
// it with Comment; a missing job fails it.
func TestBoard_Comment(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	ctx := context.Background()
	claude, _ := f.board.Users.GetByName(ctx, "claude")

	open, err := f.board.Post(ctx, f.ops, "", "@claude tidy the README")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, open.State)
	time.Sleep(50 * time.Millisecond)
	got, _ := f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteOpen, got.State, "no job: it waits")

	reply, err := f.board.Comment(ctx, claude, open.ID, "Done in commit abc123.", model.NoteDone)
	assert.NoError(t, err)
	assert.Equal(t, open.ID, reply.ParentID)
	got, _ = f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteDone, got.State)

	plain, _ := f.board.Post(ctx, f.ops, "", "a note")
	_, err = f.board.Comment(ctx, claude, plain.ID, "", model.NoteDone)
	assert.ErrorIs(t, err, note.ErrInvalid, "not an instruction")
	_, err = f.board.Comment(ctx, claude, open.ID, "x", "lost")
	assert.ErrorIs(t, err, note.ErrInvalid)
	stated, err := f.board.Comment(ctx, claude, open.ID, "", model.NoteWorking)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteWorking, stated.State)

	missing := newFixture(t, "agent/nothing")
	lost, err := missing.board.Post(ctx, missing.ops, "", "@claude go")
	assert.NoError(t, err)
	failed := missing.settled(t, lost.ID)
	assert.Equal(t, model.NoteFailed, failed.State)
	assert.Contains(t, replies(t, missing, lost.ID)[0].Body, "is not there")
}

// After a restart, an instruction left working with no run fails.
func TestBoard_Recover(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	ctx := context.Background()
	open, _ := f.board.Post(ctx, f.ops, "", "@claude go")
	assert.NoError(t, f.board.Notes.SetState(ctx, open.ID, model.NoteWorking, "gone"))
	assert.NoError(t, f.board.Recover(ctx))
	got, _ := f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteFailed, got.State)
	assert.Contains(t, replies(t, f, open.ID)[0].Body, "restarted")
}

// After a restart, an open instruction with no run starts the agent.
func TestBoard_RecoverOpen(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	note := &model.Note{Channel: model.ChannelNotes, UserID: f.ops.ID, UserName: f.ops.Name, Body: "@claude go", State: model.NoteOpen}
	assert.NoError(t, f.board.Notes.Create(ctx, note))
	assert.NoError(t, f.board.Recover(ctx))
	assert.Equal(t, model.NoteDone, f.settled(t, note.ID).State)
}

// A job that writes the reply file answers with it alone; the reply links
// the run, which keeps the output.
func TestBoard_InstructionAnswer(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	posted, err := f.board.Post(ctx, f.ops, "", "@claude answer me")
	assert.NoError(t, err)
	done := f.settled(t, posted.ID)
	assert.Equal(t, model.NoteDone, done.State)
	reply := replies(t, f, posted.ID)[0]
	assert.Equal(t, "Done: @claude answer me", reply.Body)
	assert.Equal(t, done.RunID, reply.RunID)

	failed, err := f.board.Post(ctx, f.ops, "", "@claude answer, fail")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, f.settled(t, failed.ID).State)
	assert.Equal(t, "Done: @claude answer, fail\n\nThe run ended failed (exit 1).", replies(t, f, failed.ID)[0].Body)
}

// The agent job may live in the internal folder, which the agent's user,
// a runner, does not see: the board starts it all the same.
func TestBoard_InternalJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "internal/run")
	agentFolder := *f.board.Catalogue.Find("agent")
	agentFolder.ID = model.Internal
	job := *agentFolder.Jobs[0]
	job.Folder = model.Internal
	agentFolder.Jobs = []*folder.Job{&job}
	f.board.Catalogue = catalogue{model.Internal: &agentFolder}

	posted, err := f.board.Post(context.Background(), f.ops, "", "@claude answer this")
	assert.NoError(t, err)
	done := f.settled(t, posted.ID)
	assert.Equal(t, model.NoteDone, done.State)
	reply := replies(t, f, posted.ID)
	assert.Len(t, reply, 1)
	assert.Equal(t, "Done: @claude answer this", reply[0].Body)
	run, ok := f.board.Runner.Get(done.RunID)
	assert.True(t, ok)
	assert.Equal(t, model.Internal, run.Snapshot().Folder)
}

// withPrompts sets up the fixture's board with a root with the folder proj, and
// an internal folder whose claude-prompt answers in the reply file with
// the thread, and fails a thread that ends in fail. It returns an admin.
func (f *fixture) withPrompts(t *testing.T) *model.User {
	t.Helper()
	ctx := context.Background()
	root, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)
	proj := &folder.Folder{ID: "proj", Title: "proj", Dir: filepath.Join(root, "proj")}
	assert.NoError(t, os.MkdirAll(filepath.Join(proj.Dir, "images"), 0o755))
	internal := &folder.Folder{ID: model.Internal, Title: model.Internal, Dir: root, HostJobs: true}
	job := &folder.Job{Folder: model.Internal, ID: note.PromptJob, Kind: "script", Command: []string{"sh", "-c", `printf 'In %s: %s' "$(pwd -P)" "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"; case "$TASK_UI_PROMPT" in *fail) exit 3;; esac`}}
	job.Name = note.PromptJob
	internal.Jobs = []*folder.Job{job}
	f.board.Catalogue = catalogue{"proj": proj, model.Internal: internal}
	f.board.Root = root
	admin := &model.User{Name: "admin", Role: model.RoleAdmin}
	assert.NoError(t, f.board.Users.Create(ctx, admin))
	return admin
}

// A prompt is a note of its place, open, which the prompt job answers
// with a reply from Claude; it stays off the board.
func TestBoard_Ask(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()

	asked, err := f.board.Ask(ctx, admin, "", "proj/images", "what builds here")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, asked.State)
	assert.Equal(t, "proj/images", asked.Dir)
	assert.True(t, asked.IsPrompt())
	done := f.settled(t, asked.ID)
	assert.Equal(t, model.NoteDone, done.State)
	answer := replies(t, f, asked.ID)
	assert.Len(t, answer, 1)
	assert.Equal(t, "claude", answer[0].UserName)
	assert.Equal(t, "In "+filepath.Join(f.board.Root, "proj", "images")+": what builds here", answer[0].Body)
	assert.Equal(t, done.RunID, answer[0].RunID)
	assert.Equal(t, asked.Dir, answer[0].Dir)
	assert.Equal(t, asked.WorkspaceID, answer[0].WorkspaceID)
	run, ok := f.board.Runner.Get(done.RunID)
	assert.True(t, ok)
	assert.Equal(t, admin.Name, run.Snapshot().UserName, "a prompt runs as who asks")

	board, err := f.board.Notes.List(ctx, storage.NoteFilter{})
	assert.NoError(t, err)
	assert.Len(t, board, 0, "the board lists no prompt")
	listed, err := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "proj/images"})
	assert.NoError(t, err)
	assert.Len(t, listed, 1)

	// A reply asks again with the thread, through Post as on the page.
	reply, err := f.board.Post(ctx, admin, asked.ID, "and which first")
	assert.NoError(t, err)
	assert.Equal(t, "proj/images", reply.Dir)
	assert.Eventually(t, func() bool { return len(replies(t, f, asked.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, model.NoteDone, f.settled(t, asked.ID).State)
	assert.Contains(t, replies(t, f, asked.ID)[2].Body, "admin:\nwhat builds here\n\nclaude:\nIn ")
	assert.Contains(t, replies(t, f, asked.ID)[2].Body, "admin:\nand which first")

	// The root is _.
	root, err := f.board.Ask(ctx, admin, "", model.Default, "what is here")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteDone, f.settled(t, root.ID).State)
	assert.Equal(t, "In "+f.board.Root+": what is here", replies(t, f, root.ID)[0].Body)
}

// A failed run fails the prompt, and Claude's reply says how it ended;
// only an admin asks, in a place there is.
func TestBoard_AskFailed(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()

	asked, err := f.board.Ask(ctx, admin, "", "proj", "now fail")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, f.settled(t, asked.ID).State)
	assert.Contains(t, replies(t, f, asked.ID)[0].Body, "now fail\n\nThe run ended failed (exit 3).")

	_, err = f.board.Ask(ctx, f.ops, "", "proj", "hello")
	assert.ErrorIs(t, err, note.ErrForbidden)
	_, err = f.board.Post(ctx, f.ops, asked.ID, "hello")
	assert.ErrorIs(t, err, note.ErrForbidden, "a reply in a prompt's thread is a prompt")
	_, err = f.board.Ask(ctx, admin, "", "proj/nothing", "hello")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.Ask(ctx, admin, "", "", "hello")
	assert.ErrorIs(t, err, note.ErrNoPrompt)
	_, err = f.board.Ask(ctx, admin, "", "proj", " ")
	assert.ErrorIs(t, err, note.ErrInvalid)
	plain, _ := f.board.Post(ctx, f.ops, "", "a note")
	_, err = f.board.Ask(ctx, admin, plain.ID, "", "hello")
	assert.ErrorIs(t, err, note.ErrInvalid)
}

// After a restart, an open prompt with no run starts it as its author.
func TestBoard_RecoverPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	open := &model.Note{UserID: admin.ID, UserName: admin.Name, Body: "go", State: model.NoteOpen, WorkspaceID: model.Default, Dir: "proj"}
	assert.NoError(t, f.board.Notes.Create(ctx, open))
	working := &model.Note{UserID: admin.ID, UserName: admin.Name, Body: "gone", State: model.NoteWorking, RunID: "gone", WorkspaceID: model.Default, Dir: "proj"}
	assert.NoError(t, f.board.Notes.Create(ctx, working))
	assert.NoError(t, f.board.Recover(ctx))
	assert.Equal(t, model.NoteDone, f.settled(t, open.ID).State)
	assert.Equal(t, model.NoteFailed, f.settled(t, working.ID).State)
}
