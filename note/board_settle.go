package note

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/model"
)

// dispatches are the runs of one thread the board has going, the state
// they settle its note on, and the run its note holds.
type dispatches struct {
	left  int
	state string
	run   string
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

// claims returns the run ID a thread's note holds: the first run to ask
// for it, which every run beside it leaves in place.
func (b *Board) claims(id, runID string) string {
	b.dispatchedMu.Lock()
	defer b.dispatchedMu.Unlock()
	going, ok := b.dispatched[id]
	if !ok {
		return runID
	}
	if going.run == "" {
		going.run = runID
	}
	return going.run
}

// reported folds a state a user reports into the runs of a thread the
// board has going: working while one of them goes, else the state itself.
func (b *Board) reported(id, state string) string {
	if state != model.NoteDone && state != model.NoteFailed {
		return state
	}
	b.dispatchedMu.Lock()
	defer b.dispatchedMu.Unlock()
	going, ok := b.dispatched[id]
	if !ok {
		return state
	}
	going.state = worse(going.state, state)
	return model.NoteWorking
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

// finishEach is finish with one reply per agent of the thread and a
// single settle; an answer a rerun rewrites stays finish.
func (b *Board) finishEach(ctx context.Context, note *model.Note, state, body string, again *model.Note) {
	if again != nil {
		b.finish(ctx, note, state, body, again)
		return
	}
	for _, as := range authors(b.AnswerersOf(ctx, note)) {
		b.replyAs(ctx, note, body, as)
	}
	b.settle(ctx, note, state)
}

// authors returns the agents a notice on a thread comes from, or one
// empty name, the thread's answerer, when it names no agent.
func authors(agents []string) []string {
	if len(agents) == 0 {
		return []string{""}
	}
	return agents
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
