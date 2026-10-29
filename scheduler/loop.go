package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// Catalogue finds the folder a schedule names: every folder with a job,
// the ones the workspace and project settings hide from the listings
// too, as the composition's Steps catalogue does, so a hidden folder
// keeps its schedules. folder.Discovery is one.
type Catalogue interface {
	Find(id string) *folder.Folder
}

// Starter loads a workflow with its steps, to check who may run it, and
// starts a run of it as a user. *workflow.Runner is one.
type Starter interface {
	FlowByID(ctx context.Context, id string) (*workflow.Flow, error)
	Start(ctx context.Context, wf *model.Workflow, user *model.User, input string) (*model.WorkflowRun, error)
}

// Loop starts the due schedules as detached runs, as the user each
// schedule runs as, when that user may still run the job, or every step
// of the workflow, and has the failed run of a schedule with triage on
// triaged.
type Loop struct {
	Schedules *storage.ScheduleStorage
	Users     *storage.UserStorage
	Catalogue Catalogue
	Runner    *runner.Manager
	// Flows loads the workflows the schedules of workflows name and
	// starts their runs. Nil skips such a schedule.
	Flows Starter
	// Hosts start a schedule of @host:<host> on its host. Nil skips such a
	// schedule.
	Hosts *hostrun.Starter
	// Board triages the failed run of a schedule with triage on, once
	// the run ended. Nil triages nothing.
	Board Triager
	// Cooldown is how long a schedule waits after a triage ended before
	// its next failed run gets one; zero is TriageCooldown.
	Cooldown time.Duration
	// Interval is how often the loop looks. Zero is a minute.
	Interval time.Duration
	// Now is the clock; tests move it.
	Now func() time.Time

	// triaged are the runs Triage took, by ID: one triage a run.
	// triaging are the schedules whose triage still works, by ID, and
	// cooled, by ID, when each schedule's last triage stops cooling down.
	triaged  sync.Map
	triaging sync.Map
	cooled   sync.Map
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

// Tick starts every due schedule, moves it to its next time and returns how many runs it started.
// It skips a schedule whose job or workflow is busy and disables one whose folder, job or workflow is missing.
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
	if schedule.IsWorkflow() {
		return l.fireWorkflow(ctx, schedule, now)
	}
	f, job := l.target(schedule)
	if job == nil {
		schedule.Enabled = 0
		slog.Warn("schedule disabled, its job is gone", "folder", schedule.Folder, "job", schedule.Job)
		return false, l.Schedules.Update(ctx, schedule)
	}
	if err := l.advance(ctx, schedule, now); err != nil {
		return false, err
	}
	user := l.user(ctx, schedule)
	if !l.mayRun(ctx, user, f, job) {
		slog.Warn("scheduled run skipped, its user may not run the job", "folder", schedule.Folder, "job", schedule.Job, "user", user.Name)
		return false, nil
	}
	if job.IsHost() {
		if schedule.Triages() {
			slog.Info("scheduled run not triaged, a host schedule", "job", schedule.Job)
		}
		return l.fireHosts(ctx, schedule, job, user)
	}
	session, err := l.Runner.Start(ctx, runner.Request{
		Folder:    f,
		Job:       job,
		User:      user,
		Args:      model.DecodeList(schedule.Args),
		Scheduled: true,
	})
	if errors.Is(err, runner.ErrBusy) {
		slog.Info("scheduled run skipped, the job is running", "folder", schedule.Folder, "job", schedule.Job)
		return false, nil
	}
	if err == nil && schedule.Triages() && l.Board != nil {
		go l.triageAfter(schedule, user, session)
	}
	return err == nil, err
}

// advance moves a schedule to its next time, now its last; an expression
// that fails to parse disables it.
func (l *Loop) advance(ctx context.Context, schedule *model.Schedule, now time.Time) error {
	next, err := NextRun(schedule.Expression, now)
	if err != nil {
		schedule.Enabled = 0
		return errors.Join(err, l.Schedules.Update(ctx, schedule))
	}
	schedule.LastRunAt = &now
	schedule.NextRunAt = &next
	return l.Schedules.Update(ctx, schedule)
}

// fireWorkflow starts a schedule of a workflow: a run of every step as
// the workflow's own user, the one it runs as, with no input, when that
// user may run them all; a user nobody holds skips it as one who may
// not. A workflow nobody holds disables the schedule, one with a run in
// progress skips it. A workflow run is no run of a job, so no triage
// follows it.
func (l *Loop) fireWorkflow(ctx context.Context, schedule *model.Schedule, now time.Time) (bool, error) {
	if l.Flows == nil {
		slog.Warn("scheduled run skipped, no workflows on this instance", "workflow", schedule.WorkflowID)
		return false, nil
	}
	flow, err := l.Flows.FlowByID(ctx, schedule.WorkflowID)
	if errors.Is(err, storage.ErrNotFound) {
		schedule.Enabled = 0
		slog.Warn("schedule disabled, its workflow is gone", "workflow", schedule.WorkflowID)
		return false, l.Schedules.Update(ctx, schedule)
	}
	if err != nil {
		return false, err
	}
	wf := flow.Workflow()
	if err := l.advance(ctx, schedule, now); err != nil {
		return false, err
	}
	user, err := flow.RunAs(ctx, l.user(ctx, schedule))
	if err != nil {
		slog.Warn("scheduled run skipped, its user may not run a step of the workflow", "workflow", wf.Name, "step", "", "user", wf.RunAsUserID, "err", "the user is gone")
		return false, nil
	}
	if l.Users != nil {
		if step := flow.MayRun(ctx, l.Users.Permissions(ctx, user)); step != "" {
			slog.Warn("scheduled run skipped, its user may not run a step of the workflow", "workflow", wf.Name, "step", step, "user", user.Name)
			return false, nil
		}
	}
	if schedule.Triages() {
		slog.Info("scheduled run not triaged, a workflow schedule", "workflow", wf.Name)
	}
	_, err = l.Flows.Start(ctx, wf, user, "")
	if errors.Is(err, workflow.ErrBusy) {
		slog.Info("scheduled run skipped, the workflow is running", "workflow", wf.Name)
		return false, nil
	}
	return err == nil, err
}

// fireHosts starts a schedule of @host:<host>: its command, the args,
// on the host, or on every host with an ssh login for *, a batch.
func (l *Loop) fireHosts(ctx context.Context, schedule *model.Schedule, job *folder.Job, user *model.User) (bool, error) {
	if l.Hosts == nil {
		slog.Warn("scheduled run skipped, no hosts on this instance", "job", schedule.Job)
		return false, nil
	}
	target := job.HostTarget()
	req := hostrun.Request{Targets: []string{target}, Command: strings.Join(model.DecodeList(schedule.Args), " "), Batch: target == "*", Scheduled: true}
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
