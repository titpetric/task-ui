//go:build integration

package note_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
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
	thread, err := board.Ask(ctx, ops, "", "infra", "what does this folder build")
	assert.NoError(t, err)
	settles(t, events, thread.ID)

	roots, err := runs.List(ctx, storage.RunFilter{Roots: true})
	assert.NoError(t, err)
	assert.Len(t, roots, 1)
	assert.Equal(t, model.PromptJob, roots[0].Job)
	assert.Equal(t, "", roots[0].ParentKind)
}
