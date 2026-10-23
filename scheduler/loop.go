package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue finds the folder a schedule names. folder.Discovery is one.
type Catalogue interface {
	Find(id string) *folder.Folder
}

// Loop starts the due schedules as detached runs, as the user each
// schedule runs as, when that user may still run the job.
type Loop struct {
	Schedules *storage.ScheduleStorage
	Users     *storage.UserStorage
	Catalogue Catalogue
	Runner    *runner.Manager
	// Hosts start a schedule of @host on its target. Nil skips such a
	// schedule.
	Hosts *hostrun.Starter
	// Interval is how often the loop looks. Zero is a minute.
	Interval time.Duration
	// Now is the clock; tests move it.
	Now func() time.Time
}

// NewLoop returns a loop with the defaults filled in.
func NewLoop(schedules *storage.ScheduleStorage, users *storage.UserStorage, catalogue Catalogue, manager *runner.Manager) *Loop {
	return &Loop{
		Schedules: schedules,
		Users:     users,
		Catalogue: catalogue,
		Runner:    manager,
		Interval:  time.Minute,
		Now:       time.Now,
	}
}

// Run ticks until the context ends.
func (l *Loop) Run(ctx context.Context) {
	interval := l.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := l.Tick(ctx); err != nil {
				slog.Warn("scheduler tick failed", "err", err)
			}
		}
	}
}

// Tick starts every schedule that is due and moves it to its next run.
// It returns how many runs it started. It skips a schedule whose job is
// busy until its next time, and disables one whose folder or job is
// missing, so it stops asking.
func (l *Loop) Tick(ctx context.Context) (int, error) {
	now := l.Now()
	due, err := l.Schedules.Due(ctx, now)
	if err != nil {
		return 0, err
	}
	started := 0
	var failures []error
	for i := range due {
		ok, err := l.fire(ctx, &due[i], now)
		if err != nil {
			failures = append(failures, err)
		}
		if ok {
			started++
		}
	}
	return started, errors.Join(failures...)
}

// fire starts one schedule's run and advances it.
func (l *Loop) fire(ctx context.Context, schedule *model.Schedule, now time.Time) (bool, error) {
	f, job := l.target(schedule)
	if job == nil {
		schedule.Enabled = 0
		slog.Warn("schedule disabled, its job is gone", "folder", schedule.Folder, "job", schedule.Job)
		return false, l.Schedules.Update(ctx, schedule)
	}
	next, err := NextRun(schedule.Expression, now)
	if err != nil {
		schedule.Enabled = 0
		return false, errors.Join(err, l.Schedules.Update(ctx, schedule))
	}
	schedule.LastRunAt = &now
	schedule.NextRunAt = &next
	if err := l.Schedules.Update(ctx, schedule); err != nil {
		return false, err
	}
	user := l.user(ctx, schedule)
	if !l.mayRun(ctx, user, f, job) {
		slog.Warn("scheduled run skipped, its user may not run the job", "folder", schedule.Folder, "job", schedule.Job, "user", user.Name)
		return false, nil
	}
	if job.IsHost() {
		return l.fireHosts(ctx, schedule, user)
	}
	_, err = l.Runner.Start(ctx, runner.Request{
		Folder:    f,
		Job:       job,
		User:      user,
		Scheduled: true,
	})
	if errors.Is(err, runner.ErrBusy) {
		slog.Info("scheduled run skipped, the job is running", "folder", schedule.Folder, "job", schedule.Job)
		return false, nil
	}
	return err == nil, err
}

// fireHosts starts a schedule of @host: its command, the args, on its
// target, a host or * for every host with an ssh login, a batch.
func (l *Loop) fireHosts(ctx context.Context, schedule *model.Schedule, user *model.User) (bool, error) {
	if l.Hosts == nil {
		slog.Warn("scheduled run skipped, no hosts on this instance", "job", schedule.Job, "target", schedule.Target)
		return false, nil
	}
	req := hostrun.Request{Targets: []string{schedule.Target}, Command: strings.Join(model.DecodeList(schedule.Args), " "), Batch: schedule.Target == "*", Scheduled: true}
	_, _, err := l.Hosts.Start(ctx, user, req)
	return err == nil, err
}

// target resolves the folder and job a schedule names. A shell job is
// none: it runs with an admin at the keyboard, never on a schedule.
func (l *Loop) target(schedule *model.Schedule) (*folder.Folder, *folder.Job) {
	f := l.Catalogue.Find(schedule.Folder)
	if f == nil {
		return nil, nil
	}
	job := f.Job(schedule.Job)
	if job.IsShell() {
		return f, nil
	}
	return f, job
}

// mayRun reports whether the schedule's user may run its job, as they
// would by hand: they see the folder, and the run permission and the
// allow lists let them. Without the users, the loop runs every job.
func (l *Loop) mayRun(ctx context.Context, user *model.User, f *folder.Folder, job *folder.Job) bool {
	if l.Users == nil {
		return true
	}
	set := l.Users.Permissions(ctx, user)
	if job.IsHost() {
		return set.Allowed("hosts")
	}
	return set.Run(f.ID, f.Allow, job.Allow)
}

// user resolves whom to attribute the run to: the schedule's user when
// the row still exists, otherwise the name the schedule kept.
func (l *Loop) user(ctx context.Context, schedule *model.Schedule) *model.User {
	if l.Users != nil && schedule.UserID != "" {
		if user, err := l.Users.Get(ctx, schedule.UserID); err == nil {
			return user
		}
	}
	return &model.User{ID: schedule.UserID, Name: schedule.UserName}
}
