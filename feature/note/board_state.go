package note

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// setState records a state a user reports. Working from someone over
// MCP lets go of a run of the board's that is over: the note keeps the
// run it had from an earlier answer, such as the agent's before a reply
// handed the thread to another user, and only a live run is the board's
// to wait on.
//
// An agent of the thread reporting an end leaves the note working for
// as long as a second run goes, as reported folds it in: the last run
// settles it. Anyone else ends the thread where they say, because a run
// that hangs would otherwise hold the note open for good.
func (b *Board) setState(ctx context.Context, user *model.User, note *model.Note, state string) error {
	if state == model.NoteWorking && note.RunID != "" && !b.live(note.RunID) {
		if err := b.Notes.Release(ctx, note.ID, state); err != nil {
			return err
		}
		note.State, note.RunID = state, ""
		return nil
	}
	if b.byAgent(ctx, user, note) {
		state = b.reported(note.ID, state)
	} else if state == model.NoteDone || state == model.NoteFailed {
		b.resolve(note.ID)
	}
	if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
		return err
	}
	note.State = state
	return nil
}

// byAgent reports whether a state comes from an agent the thread names,
// whose report waits for the thread's other agents.
func (b *Board) byAgent(ctx context.Context, user *model.User, note *model.Note) bool {
	if user == nil || b.Users == nil {
		return false
	}
	for _, agent := range b.threadAgents(ctx, note) {
		if agent.Name == user.Name {
			return true
		}
	}
	return false
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

// Recover fails the instructions a restart cut short: a note working on a run of the board's has no run to wait on.
// The notice comes from every agent of the thread, a reply each, because a run of each was cut.
// A note working with no run is an external agent's, which goes on.
// An open instruction to the agent with no run, never started or set open by hand, starts every agent the thread names.
// An open prompt with no run starts its prompt.
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
		if open[i].RunID != "" {
			continue
		}
		b.dispatchFor(&open[i], b.threadAgents(ctx, &open[i]), nil)
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
		b.finishEach(ctx, &note, model.NoteFailed, "task-ui restarted while the run was going.", b.working(ctx, &note))
	}
	return nil
}
