package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// TriageCooldown is how long a schedule waits after a triage ended
// before its next failed run gets one, when the loop sets none.
const TriageCooldown = time.Hour

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
// killed, as the schedule's user, once a run, one at a time a schedule
// and at most once a Cooldown a schedule, and returns the verdict; ""
// for a run it leaves alone: one that ended any other way, a run of a
// schedule without triage, a host's or Claude's own, one triaged before,
// one while the schedule's last triage still works or cools down, or one
// the board refuses, each logged. A stop before the triage ends is its
// error.
func (l *Loop) Triage(ctx context.Context, schedule *model.Schedule, user *model.User, run *model.Run) (string, error) {
	switch {
	case l.Board == nil || !schedule.Triages() || run == nil || user == nil:
		return "", nil
	case run.State != string(model.StateFailed) && run.State != string(model.StateKilled):
		return "", nil
	case run.IsHost() || run.IsClaude():
		return "", nil
	}
	if until, ok := l.cooled.Load(schedule.ID); ok && l.Now().Before(until.(time.Time)) {
		slog.Info("scheduled run not triaged, the schedule's last triage cools down", "run", run.ID, "folder", schedule.Folder, "job", schedule.Job, "until", until.(time.Time).Format(time.RFC3339))
		return "", nil
	}
	if _, busy := l.triaging.LoadOrStore(schedule.ID, true); busy {
		slog.Info("scheduled run not triaged, the schedule's last triage still works", "run", run.ID, "folder", schedule.Folder, "job", schedule.Job)
		return "", nil
	}
	defer l.triaging.Delete(schedule.ID)
	if _, taken := l.triaged.LoadOrStore(run.ID, true); taken {
		return "", nil
	}
	_, verdict, err := l.Board.TriageRun(ctx, user, run.ID)
	l.cooled.Store(schedule.ID, l.Now().Add(l.cooldown()))
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

// cooldown is the loop's Cooldown, TriageCooldown when unset.
func (l *Loop) cooldown() time.Duration {
	if l.Cooldown > 0 {
		return l.Cooldown
	}
	return TriageCooldown
}
