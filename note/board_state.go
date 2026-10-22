package note

import (
	"context"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// setState records a state a user reports. Working from someone over
// MCP lets go of a run of the board's that is over: the note keeps the
// run it had from an earlier answer, such as the agent's before a reply
// handed the thread to another user, and only a live run is the board's
// to wait on.
func (b *Board) setState(ctx context.Context, note *model.Note, state string) error {
	if state == model.NoteWorking && note.RunID != "" && !b.live(note.RunID) {
		if err := b.Notes.Release(ctx, note.ID, state); err != nil {
			return err
		}
		note.State, note.RunID = state, ""
		return nil
	}
	if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
		return err
	}
	note.State = state
	return nil
}

// live reports whether a run of the board's is going, or still settling.
func (b *Board) live(runID string) bool {
	if b.Runner == nil {
		return false
	}
	session, ok := b.Runner.Get(runID)
	if !ok {
		return false
	}
	select {
	case <-session.Done():
		return false
	default:
		return true
	}
}

// Recover fails the instructions a restart cut short: a note still
// working on a run of the board's has no run to wait on any more. A note
// working with no run is an external agent's, which goes on. An open
// instruction to the agent with no run yet, one that never started or one
// set open by hand, starts the agent, and an open prompt with no run
// starts its prompt.
func (b *Board) Recover(ctx context.Context) error {
	open, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteOpen})
	if err != nil {
		return err
	}
	prompts, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteOpen, Prompts: true})
	if err != nil {
		return err
	}
	for i := range open {
		if open[i].RunID == "" && b.threadForAgent(ctx, &open[i]) {
			b.dispatch(&open[i], nil)
		}
	}
	for i := range prompts {
		if prompts[i].RunID == "" {
			b.reask(ctx, &prompts[i])
		}
	}
	working, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteWorking})
	if err != nil {
		return err
	}
	workingPrompts, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteWorking, Prompts: true})
	if err != nil {
		return err
	}
	for _, note := range append(working, workingPrompts...) {
		if note.RunID == "" && !note.IsPrompt() {
			continue
		}
		if note.RunID != "" && b.Runner != nil {
			if _, live := b.Runner.Get(note.RunID); live {
				continue
			}
		}
		b.finish(ctx, &note, model.NoteFailed, "task-ui restarted while the run was going.", b.working(ctx, &note))
	}
	return nil
}
