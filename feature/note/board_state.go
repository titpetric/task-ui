package note

import (
	"context"
	"log/slog"
	"os"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/claudesession"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// setState records a state a user reports, working letting go of a run of the board's that is over.
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

// byAgent reports whether a state comes from an agent the thread names, which waits for the thread's other agents.
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

// Recover fails the instructions a restart cut short, resumes a prompt's session, and starts what never ran.
// The notice comes from every agent of the thread, a reply each, because a run of each was cut.
// A note working without a run is an external agent's, which goes on.
// An open instruction to the agent, never started or set open by hand, starts every agent the thread names.
// An open prompt starts its prompt.
func (b *Board) Recover(ctx context.Context) error {
	open, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteOpen})
	if err != nil {
		return err
	}
	prompts, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteOpen, Prompts: true})
	if err != nil {
		return err
	}
	working, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteWorking})
	if err != nil {
		return err
	}
	workingPrompts, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteWorking, Prompts: true})
	if err != nil {
		return err
	}
	keep, seen := b.liveSessions(), map[string]bool{}
	var resumes []pendingResume
	for _, note := range append(working, workingPrompts...) {
		if note.RunID == "" && !note.IsPrompt() {
			continue
		}
		if note.RunID != "" && b.Runner != nil {
			if _, live := b.Runner.Get(note.RunID); live {
				continue
			}
		}
		if note.IsPrompt() {
			if resume, ok := b.resumePrompt(ctx, &note, seen, keep); ok {
				resumes = append(resumes, resume)
				continue
			}
		}
		b.finishEach(ctx, &note, model.NoteFailed, "task-ui restarted while the run was going.", b.working(ctx, &note))
	}
	// Before any run starts: a directory a re-asked prompt is about to make is not an orphan.
	b.cleanSessions(keep)
	for _, resume := range resumes {
		b.dispatchPrompt(resume.note, nil, resume.req, nil)
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
	return nil
}

// pendingResume is a working prompt and the request that continues its session.
type pendingResume struct {
	note *model.Note
	req  runner.Request
}

// resumePrompt prepares a working prompt's Claude session to continue, and answers whether it can.
func (b *Board) resumePrompt(ctx context.Context, note *model.Note, seen, keep map[string]bool) (pendingResume, bool) {
	if note.RunID == "" || seen[note.RunID] || b.Runs == nil || b.Runner == nil || b.Users == nil || b.ClaudeDir == "" {
		return pendingResume{}, false
	}
	old, err := b.Runs.Get(ctx, note.RunID)
	if err != nil || old.ClaudeSession == "" {
		return pendingResume{}, false
	}
	if _, err := os.Stat(claudesession.Dir(b.ClaudeDir, old.UserName, old.ClaudeSession)); err != nil {
		return pendingResume{}, false
	}
	user, err := b.Users.Get(ctx, note.UserID)
	if err != nil {
		return pendingResume{}, false
	}
	req, err := b.promptRequest(ctx, user, note.Dir, note.AboutWorkflowRunID != "")
	if err != nil {
		return pendingResume{}, false
	}
	req.ClaudeSession, req.Resume = old.ClaudeSession, true
	seen[note.RunID] = true
	keep[old.ClaudeSession] = true
	return pendingResume{note: note, req: req}, true
}

// liveSessions returns the Claude sessions the runner holds, for the cleanup to keep.
func (b *Board) liveSessions() map[string]bool {
	keep := map[string]bool{}
	if b.Runner == nil {
		return keep
	}
	for _, session := range b.Runner.List() {
		if session.Run.ClaudeSession != "" {
			keep[session.Run.ClaudeSession] = true
		}
	}
	return keep
}

// stopped reports a shutdown's kill, which leaves the session for Recover, else removes the run's session.
func (b *Board) stopped(req runner.Request) bool {
	if b.closing.Load() {
		return true
	}
	if b.ClaudeDir != "" && req.User != nil {
		_ = claudesession.Remove(claudesession.Dir(b.ClaudeDir, req.User.Name, req.ClaudeSession))
	}
	return false
}

// Close stops the board settling the runs a shutdown kills.
func (b *Board) Close() { b.closing.Store(true) }

// cleanSessions removes the session directories no live or resumed run names.
func (b *Board) cleanSessions(keep map[string]bool) {
	if b.ClaudeDir == "" {
		return
	}
	if err := claudesession.Clean(b.ClaudeDir, keep); err != nil {
		slog.Warn("claude sessions not cleaned", "err", err)
	}
}
