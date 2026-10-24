package scheduler

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// Triager triages a failed run of a schedule as the schedule's user and
// leaves the note of it: Claude on the run's page, then the board.
// note.Board is one.
type Triager interface {
	TriageRun(stop context.Context, user *model.User, runID string) (*model.Note, string, error)
}

// triageAfter waits for a schedule's run to end, then has it triaged
// when it failed.
func (l *Loop) triageAfter(schedule *model.Schedule, user *model.User, session *runner.Session) {
	<-session.Done()
	run := session.Snapshot()
	run.State = string(session.Status().State)
	if _, err := l.Triage(context.Background(), schedule, user, &run); err != nil {
		slog.Warn("scheduled run not triaged", "run", run.ID, "folder", schedule.Folder, "job", schedule.Job, "err", err)
	}
}

// Triage has the Board triage a run of a schedule that ended failed or
// killed, as the schedule's user, once a run, and returns the verdict;
// "" for a run it leaves alone: one that ended any other way, a run of
// a schedule without triage, a host's or Claude's own, one triaged
// before, or one the board refuses, which it logs. A stop before the
// triage ends is its error.
func (l *Loop) Triage(ctx context.Context, schedule *model.Schedule, user *model.User, run *model.Run) (string, error) {
	switch {
	case l.Board == nil || !schedule.Triages() || run == nil || user == nil:
		return "", nil
	case run.State != string(model.StateFailed) && run.State != string(model.StateKilled):
		return "", nil
	case run.IsHost() || run.IsClaude():
		return "", nil
	}
	if _, taken := l.triaged.LoadOrStore(run.ID, true); taken {
		return "", nil
	}
	_, verdict, err := l.Board.TriageRun(ctx, user, run.ID)
	switch {
	case ctx.Err() != nil:
		return "", ctx.Err()
	case err != nil:
		slog.Warn("scheduled run not triaged", "run", run.ID, "folder", schedule.Folder, "job", schedule.Job, "user", user.Name, "err", err)
		return "", nil
	}
	slog.Info("scheduled run triaged", "run", run.ID, "folder", schedule.Folder, "job", schedule.Job, "user", user.Name, "result", verdict)
	return verdict, nil
}
