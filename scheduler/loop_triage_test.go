package scheduler_test

import (
	"context"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/scheduler"
)

// board is the Triager of the tests: it records the triages asked of
// it and answers each with its verdict, or its error.
type board struct {
	mu      sync.Mutex
	runs    []string
	users   []string
	verdict string
	err     error
}

func (b *board) TriageRun(_ context.Context, user *model.User, runID string) (*model.Note, string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.runs = append(b.runs, runID)
	b.users = append(b.users, user.Name)
	if b.err != nil {
		return nil, "", b.err
	}
	return &model.Note{Body: "noted"}, b.verdict, nil
}

// The board of the notes is the loop's Triager.
func TestTriager(t *testing.T) {
	t.Parallel()
	var triager scheduler.Triager = (*note.Board)(nil)
	assert.NotNil(t, &triager)
	triager = &board{}
	assert.NotNil(t, triager)
}

func (b *board) triaged() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string{}, b.runs...)
}

// A failed run of a schedule with triage on gets one triage, as the
// schedule's user; a run that ended well, a schedule without triage, a
// host's or Claude's run, and a loop without a board triage nothing.
func TestLoop_Triage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := &board{verdict: note.TriageResolved}
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	loop.Board = b
	ops := &model.User{ID: "u1", Name: "ops"}
	on := &model.Schedule{Folder: "infra", Job: "plan", Triage: 1}
	failed := &model.Run{ID: "r1", Folder: "infra", Job: "plan", State: string(model.StateFailed)}

	verdict, err := loop.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict)
	assert.Equal(t, []string{"r1"}, b.triaged())
	assert.Equal(t, []string{"ops"}, b.users)

	verdict, err = loop.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "once a run")
	assert.Equal(t, []string{"r1"}, b.triaged())

	killed := &model.Run{ID: "r2", Folder: "infra", Job: "plan", State: string(model.StateKilled)}
	verdict, err = loop.Triage(ctx, on, ops, killed)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict, "a killed run is a failure")
	assert.Equal(t, []string{"r1", "r2"}, b.triaged())

	left := []struct {
		name     string
		schedule *model.Schedule
		run      *model.Run
	}{
		{"ok", on, &model.Run{ID: "r3", Folder: "infra", Job: "plan", State: string(model.StateOK)}},
		{"stopped", on, &model.Run{ID: "r4", Folder: "infra", Job: "plan", State: string(model.StateStopped)}},
		{"triage off", &model.Schedule{Folder: "infra", Job: "plan"}, &model.Run{ID: "r5", Folder: "infra", Job: "plan", State: string(model.StateFailed)}},
		{"a host", &model.Schedule{Folder: "_", Job: "@host:web", Triage: 1}, &model.Run{ID: "r6", Folder: "_", Job: "@host:web", Target: "web", State: string(model.StateFailed)}},
		{"Claude's own", on, &model.Run{ID: "r7", Folder: "infra", Job: model.ClaudeJob, State: string(model.StateFailed)}},
		{"no run", on, nil},
	}
	for _, c := range left {
		verdict, err := loop.Triage(ctx, c.schedule, ops, c.run)
		assert.NoError(t, err, c.name)
		assert.Equal(t, "", verdict, c.name)
	}
	assert.Equal(t, []string{"r1", "r2"}, b.triaged())

	off := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	verdict, err = off.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "no board")
}

// A board that refuses the triage, for a user without prompt/edit or a
// run no prompt reaches, ends it with a log line and no error; a stop
// ends it with the stop's error.
func TestLoop_TriageRefused(t *testing.T) {
	t.Parallel()
	ops := &model.User{ID: "u1", Name: "ops"}
	on := &model.Schedule{Folder: "infra", Job: "plan", Triage: 1}
	failed := &model.Run{ID: "r1", Folder: "infra", Job: "plan", State: string(model.StateFailed)}

	for _, refusal := range []error{note.ErrForbidden, note.ErrNoPrompt, note.ErrPromptFailed} {
		b := &board{err: refusal}
		loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
		loop.Board = b
		verdict, err := loop.Triage(context.Background(), on, ops, failed)
		assert.NoError(t, err, refusal.Error())
		assert.Equal(t, "", verdict)
		assert.Equal(t, []string{"r1"}, b.triaged(), "asked, and refused")
	}

	stopped, cancel := context.WithCancel(context.Background())
	cancel()
	b := &board{err: context.Canceled}
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	loop.Board = b
	_, err := loop.Triage(stopped, on, ops, failed)
	assert.ErrorIs(t, err, context.Canceled)
}
