//go:build integration

package note_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/feature/orb/recording"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/claudesession"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// folders is a catalogue of one folder.
type folders []*folder.Folder

func (c folders) List() []*folder.Folder { return c }

func (c folders) Find(id string) *folder.Folder {
	for _, f := range c {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// askingBoard is a board that starts real prompt runs: the folder infra
// under its root, and true as the agent command, so a prompt's run ends
// ok without Claude.
func askingBoard(t *testing.T) (*note.Board, *storage.RunStorage, *model.User) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users, runs := storage.NewUserStorage(db), storage.NewRunStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, ops))
	root := t.TempDir()
	assert.NoError(t, os.Mkdir(filepath.Join(root, "infra"), 0o755))
	infra := &folder.Folder{ID: "infra", Title: "infra", Dir: filepath.Join(root, "infra")}
	manager := runner.New(runner.Options{Runs: runs, Recordings: &recording.Store{Root: t.TempDir()}})
	board := &note.Board{
		Notes: storage.NewNoteStorage(db), Users: users, Runs: runs,
		Catalogue: folders{infra}, Runner: manager, Events: event.NewBus(),
		Root: root, AgentCommand: []string{"true"},
	}
	return board, runs, ops
}

// settles waits for a thread to leave the working state, which the board
// publishes when its run ended.
func settles(t *testing.T, events <-chan event.Event, id string) {
	t.Helper()
	for {
		select {
		case e := <-events:
			row, ok := e.Data.(model.Note)
			if ok && row.ID == id && row.State != model.NoteOpen && row.State != model.NoteWorking {
				return
			}
		case <-time.After(10 * time.Second):
			t.Fatal("the prompt did not settle")
		}
	}
}

// A prompt about a run sits under that run, so the rail lists the run
// alone and the run's page lists the prompt below it.
func TestBoard_AskRunParent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, ops := askingBoard(t)
	job := &model.Run{Folder: "infra", Job: "build", UserID: ops.ID, UserName: ops.Name}
	assert.NoError(t, runs.Create(ctx, job))

	events, cancel := board.Events.Subscribe(note.Topic)
	defer cancel()
	thread, err := board.AskRun(ctx, ops, job.ID, "why did this run end the way it did")
	assert.NoError(t, err)
	settles(t, events, thread.ID)

	children, err := runs.Children(ctx, job.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, model.PromptJob, children[0].Job)
	assert.Equal(t, storage.ParentRun, children[0].ParentKind)

	roots, err := runs.List(ctx, storage.RunFilter{Roots: true})
	assert.NoError(t, err)
	assert.Equal(t, []string{job.ID}, []string{roots[0].ID})
	assert.Len(t, roots, 1, "the rail lists the run and not the prompt it was asked about")
}

// A prompt a person asked of a folder sits under nothing: it is a row of
// the rail itself.
func TestBoard_AskParent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, ops := askingBoard(t)

	events, cancel := board.Events.Subscribe(note.Topic)
	defer cancel()
	heard := note.Spoken{Audio: "voice/2026-11-11T10-00-00.wav", Text: "what does this folder build"}
	thread, err := board.Ask(ctx, ops, "", "infra", "what does this folder build", heard)
	assert.NoError(t, err)
	settles(t, events, thread.ID)
	assert.Equal(t, heard.Audio, thread.AudioPath, "the recording stays with the message")
	assert.Equal(t, heard.Text, thread.Spoken)

	roots, err := runs.List(ctx, storage.RunFilter{Roots: true})
	assert.NoError(t, err)
	assert.Len(t, roots, 1)
	assert.Equal(t, model.PromptJob, roots[0].Job)
	assert.Equal(t, "", roots[0].ParentKind)
}

// A prompt answers as the agent it names, and the thread keeps it.
func TestBoard_AskAsAgent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, ops := askingBoard(t)
	agent := &model.User{Name: "consuela", Role: model.RoleViewer, AgentReply: true, Mcp: true}
	assert.NoError(t, board.Users.Create(ctx, agent))

	events, cancel := board.Events.Subscribe(note.Topic)
	defer cancel()
	thread, err := board.AskAs(ctx, ops, "", "infra", "who broke the build", note.Spoken{}, "consuela", "folder")
	assert.NoError(t, err)
	assert.Equal(t, "consuela", thread.Agent)
	assert.Equal(t, "folder", thread.Origin)
	assert.Equal(t, "consuela", board.AnswererOf(ctx, thread))
	settles(t, events, thread.ID)

	// The answer was made when the run started, and its body filled that reply.
	thread, err = board.Notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	replies, err := board.Notes.Replies(ctx, []string{thread.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[thread.ID], 1)
	answer := replies[thread.ID][0]
	assert.Equal(t, thread.ID, answer.ParentID)
	assert.Equal(t, "consuela", answer.Agent)
	assert.Equal(t, thread.RunID, answer.RunID)
	assert.True(t, strings.TrimSpace(answer.Body) != "", "the answer has its body")
}

// A prompt with no agent names the first agent, and refuses a user who is none.
func TestBoard_AskAsAgentDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, ops := askingBoard(t)
	agent := &model.User{Name: "consuela", Role: model.RoleViewer, AgentReply: true, Mcp: true}
	plain := &model.User{Name: "plain", Role: model.RoleViewer}
	assert.NoError(t, board.Users.Create(ctx, agent))
	assert.NoError(t, board.Users.Create(ctx, plain))

	_, err := board.AskAs(ctx, ops, "", "infra", "hello", note.Spoken{}, "plain", "folder")
	assert.True(t, errors.Is(err, note.ErrInvalid), "a user without agent reply is no agent")

	events, cancel := board.Events.Subscribe(note.Topic)
	defer cancel()
	thread, err := board.AskAs(ctx, ops, "", "infra", "hello", note.Spoken{}, "", "folder")
	assert.NoError(t, err)
	assert.Equal(t, "consuela", thread.Agent)
	settles(t, events, thread.ID)
}

// A workflow page's prompt answers as the agent it names.
func TestBoard_AskAreaAs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, ops := askingBoard(t)
	agent := &model.User{Name: "consuela", Role: model.RoleViewer, AgentReply: true, Mcp: true}
	assert.NoError(t, board.Users.Create(ctx, agent))

	events, cancel := board.Events.Subscribe(note.Topic)
	defer cancel()
	thread, err := board.AskAreaAs(ctx, ops, model.AreaWorkflows, "infra", "build a workflow", "consuela")
	assert.NoError(t, err)
	assert.Equal(t, "consuela", thread.Agent)
	assert.Equal(t, model.AreaWorkflows, thread.Area)
	settles(t, events, thread.ID)
}

// resumingBoard is a board with a Claude session directory and an agent command that writes its environment to a file.
func resumingBoard(t *testing.T) (*note.Board, *storage.RunStorage, *storage.NoteStorage, *model.User, string, string) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users, runs, notes := storage.NewUserStorage(db), storage.NewRunStorage(db), storage.NewNoteStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, ops))

	root, claudeDir := t.TempDir(), t.TempDir()
	assert.NoError(t, os.Mkdir(filepath.Join(root, "infra"), 0o755))
	seen := filepath.Join(t.TempDir(), "env")
	manager := runner.New(runner.Options{Runs: runs, Recordings: &recording.Store{Root: t.TempDir()}, ClaudeDir: claudeDir})
	board := &note.Board{
		Notes: notes, Users: users, Runs: runs,
		Catalogue: folders{{ID: "infra", Title: "infra", Dir: filepath.Join(root, "infra")}},
		Runner:    manager, Root: root, ClaudeDir: claudeDir, Events: event.NewBus(),
		AgentCommand: []string{"sh", "-c", "env > " + seen},
	}
	return board, runs, notes, ops, claudeDir, seen
}

// A working prompt whose Claude session is on disk resumes that session under a new run.
func TestBoard_RecoverResumesPrompt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, notes, ops, claudeDir, seen := resumingBoard(t)

	id, orphan := model.NewID(), model.NewID()
	old := &model.Run{ID: model.NewID(), Folder: "infra", Job: model.PromptJob, UserName: ops.Name, UserID: ops.ID, State: model.StateKilled, ClaudeSession: id}
	assert.NoError(t, runs.Create(ctx, old))
	assert.NoError(t, os.MkdirAll(claudesession.Dir(claudeDir, ops.Name, id), 0o700))
	assert.NoError(t, os.MkdirAll(claudesession.Dir(claudeDir, ops.Name, orphan), 0o700))

	thread := &model.Note{UserName: ops.Name, UserID: ops.ID, Body: "do the work", WorkspaceID: "infra", Dir: "infra", State: model.NoteWorking, RunID: old.ID}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.NoError(t, board.Recover(ctx))

	env := waitFile(t, seen)
	assert.Contains(t, env, "TASK_UI_CLAUDE_SESSION="+id+"\n")
	assert.Contains(t, env, "TASK_UI_CLAUDE_RESUME=1\n")
	assert.Contains(t, env, "TASK_UI_CLAUDE_CONFIG="+claudesession.Dir(claudeDir, ops.Name, id)+"\n")

	got, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.True(t, got.RunID != old.ID)
	waitGone(t, claudesession.Dir(claudeDir, ops.Name, id))
	_, err = os.Stat(claudesession.Dir(claudeDir, ops.Name, orphan))
	assert.True(t, os.IsNotExist(err), "a session no run holds is removed")
}

// A shutdown leaves a prompt's note working and its session on disk, for Recover.
func TestBoard_CloseKeepsSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, notes, ops, claudeDir, seen := resumingBoard(t)
	board.AgentCommand = []string{"sh", "-c", "env > " + seen + "; sleep 0.2"}

	thread, err := board.Ask(ctx, ops, "", "infra", "do the work", note.Spoken{})
	assert.NoError(t, err)
	waitFile(t, seen)
	got, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	run, err := runs.Get(ctx, got.RunID)
	assert.NoError(t, err)
	dir := claudesession.Dir(claudeDir, ops.Name, run.ClaudeSession)
	assert.NoError(t, os.MkdirAll(dir, 0o700))

	board.Close()
	for i := 0; i < 20; i++ {
		if _, live := board.Runner.Get(run.ID); !live {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	got, err = notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteWorking, got.State, "the restart resumes the note")
	_, err = os.Stat(dir)
	assert.NoError(t, err, "the session stays for the restart")
}

// waitGone waits for a path to be removed.
func waitGone(t *testing.T, path string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s is still there", path)
}

// A working prompt with no session on disk fails with the restart notice.
func TestBoard_RecoverNoSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, notes, ops, _, _ := resumingBoard(t)

	old := &model.Run{ID: model.NewID(), Folder: "infra", Job: model.PromptJob, UserName: ops.Name, UserID: ops.ID, State: model.StateKilled, ClaudeSession: model.NewID()}
	assert.NoError(t, runs.Create(ctx, old))
	thread := &model.Note{UserName: ops.Name, UserID: ops.ID, Body: "do the work", WorkspaceID: "infra", Dir: "infra", State: model.NoteWorking, RunID: old.ID}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.NoError(t, board.Recover(ctx))

	got, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, got.State)
}

// waitFile waits for a file and returns its contents.
func waitFile(t *testing.T, path string) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		if data, err := os.ReadFile(path); err == nil && strings.Contains(string(data), "TASK_UI_CLAUDE_SESSION") {
			return string(data)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no environment at %s", path)
	return ""
}
