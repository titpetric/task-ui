package note

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/model"
)

// dispatches are the runs of one thread the board has going, and the
// state they settle its note on.
type dispatches struct {
	left  int
	state string
}

// dispatchFor starts the Notes agent for each agent of a thread, a run
// each; the last run to end settles the note. again returns the reply
// an agent rewrites, nil for a new reply.
func (b *Board) dispatchFor(thread *model.Note, agents []*model.User, again func(*model.User) *model.Note) {
	b.dispatching(thread.ID, len(agents))
	for _, agent := range agents {
		var rewrite *model.Note
		if again != nil {
			rewrite = again(agent)
		}
		b.dispatch(thread, rewrite, agent)
	}
}

// dispatching counts n runs of a thread before the first starts. A run
// that ends while the count stands leaves the note working.
func (b *Board) dispatching(id string, n int) {
	if n < 1 {
		return
	}
	b.dispatchedMu.Lock()
	defer b.dispatchedMu.Unlock()
	if b.dispatched == nil {
		b.dispatched = map[string]*dispatches{}
	}
	if going, ok := b.dispatched[id]; ok {
		going.left += n
		return
	}
	b.dispatched[id] = &dispatches{left: n}
}

// settling returns the state one run leaves a thread's note in: working
// while another run of it goes, else the worst state reported. A thread
// under no count settles on the state it reports.
func (b *Board) settling(id, state string) string {
	b.dispatchedMu.Lock()
	defer b.dispatchedMu.Unlock()
	going, ok := b.dispatched[id]
	if !ok {
		return state
	}
	going.left, going.state = going.left-1, worse(going.state, state)
	if going.left > 0 {
		return model.NoteWorking
	}
	delete(b.dispatched, id)
	return going.state
}

// worse returns the state two runs settle a thread on where they
// disagree: failed, because a done hides the failure.
func worse(state, next string) string {
	if state == model.NoteFailed || next == model.NoteFailed {
		return model.NoteFailed
	}
	return next
}

// settle records the state an instruction ended in, as settling reads it.
func (b *Board) settle(ctx context.Context, note *model.Note, state string) {
	state = b.settling(note.ID, state)
	if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
		slog.Warn("note state not saved", "note", note.ID, "err", err)
	}
	note.State = state
	b.publish(note)
}
