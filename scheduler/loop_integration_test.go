//go:build integration

package scheduler_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

type fixture struct {
	loop      *scheduler.Loop
	schedules *storage.ScheduleStorage
	runs      *storage.RunStorage
	users     *storage.UserStorage
	workflows *storage.WorkflowStorage
	// flows records the workflow runs the loop starts.
	flows *fakeStarter
	now   time.Time
	// ops is the runner every schedule of the fixture runs as.
	ops *model.User
}

// fakeStarter records the workflow runs the loop starts, with the real
// runner's MayRun; busy answers every start with ErrBusy.
type fakeStarter struct {
	*workflow.Runner
	mu     sync.Mutex
	starts []string
	busy   bool
}

func (s *fakeStarter) Start(_ context.Context, wf *model.Workflow, user *model.User, input string) (*model.WorkflowRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return nil, workflow.ErrBusy
	}
	s.starts = append(s.starts, wf.Name+" as "+user.Name+" with "+strconv.Quote(input))
	return &model.WorkflowRun{ID: model.NewID(), WorkflowID: wf.ID}, nil
}

func (s *fakeStarter) started() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.starts...)
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "infra"), 0o755))
	infra := &folder.Folder{ID: "infra", Dir: filepath.Join(dir, "infra")}
	plan := &folder.Job{Folder: "infra", ID: "plan", Command: []string{"sh", "-c", "echo planned"}}
	plan.Name = "plan"
	slow := &folder.Job{Folder: "infra", ID: "slow", Command: []string{"sh", "-c", "sleep 30"}}
	slow.Name = "slow"
	greet := &folder.Job{Folder: "infra", ID: "greet", Command: []string{"sh", "-c", `echo "hi $*"`, "sh"}}
	greet.Name, greet.TakesArgs = "greet", true
	infra.Jobs = []*folder.Job{plan, slow, greet}

	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{
		Runs:       runs,
		Recordings: &recording.Store{Root: t.TempDir()},
		Grace:      time.Second,
	})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	schedules := storage.NewScheduleStorage(db)
	users := storage.NewUserStorage(db)
	folders := catalogue{"infra": infra, model.Default: {ID: model.Default, Dir: dir}}
	loop := scheduler.NewLoop(schedules, users, folders, manager)
	workflows := storage.NewWorkflowStorage(db)
	flows := &fakeStarter{Runner: workflow.New(workflows, storage.NewWorkflowRunStorage(db), folders, manager)}
	loop.Workflows, loop.Flows = workflows, flows
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(context.Background(), ops))
	f := &fixture{loop: loop, schedules: schedules, runs: runs, users: users, workflows: workflows, flows: flows, ops: ops, now: time.Date(2026, time.October, 17, 10, 0, 0, 0, time.UTC)}
	loop.Now = func() time.Time { return f.now }
	return f
}

func (f *fixture) schedule(t *testing.T, job, expression string) *model.Schedule {
	t.Helper()
	due := f.now.Add(-time.Minute)
	schedule := &model.Schedule{Folder: "infra", Job: job, Expression: expression, Enabled: 1, NextRunAt: &due, UserID: f.ops.ID, UserName: "ops"}
	assert.NoError(t, f.schedules.Create(context.Background(), schedule))
	return schedule
}

// workflow makes a workflow of infra jobs and a due schedule of it as ops.
func (f *fixture) workflow(t *testing.T, name string, jobs ...string) (*model.Workflow, *model.Schedule) {
	t.Helper()
	steps := make([]model.WorkflowStep, 0, len(jobs))
	for _, job := range jobs {
		steps = append(steps, model.WorkflowStep{Folder: "infra", Job: job})
	}
	wf := &model.Workflow{Name: name, UserID: f.ops.ID, UserName: "ops"}
	assert.NoError(t, f.workflows.Create(context.Background(), wf, steps))
	due := f.now.Add(-time.Minute)
	schedule := &model.Schedule{WorkflowID: wf.ID, Expression: "0 * * * *", Enabled: 1, NextRunAt: &due, UserID: f.ops.ID, UserName: "ops"}
	assert.NoError(t, f.schedules.Create(context.Background(), schedule))
	return wf, schedule
}

// A due schedule of a workflow starts a run of it as the schedule's user
// with no input, and moves to its next time.
func TestLoop_TickWorkflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	_, schedule := f.workflow(t, "release", "plan")

	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)
	assert.Equal(t, []string{`release as ops with ""`}, f.flows.started())
	got, err := f.schedules.Get(ctx, schedule.ID)
	assert.NoError(t, err)
	assert.Equal(t, f.now, got.LastRunAt.UTC())
	assert.Equal(t, f.now.Add(time.Hour), got.NextRunAt.UTC())
	assert.True(t, got.IsEnabled())

	// Not due again until then.
	started, err = f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	assert.Len(t, f.flows.started(), 1)
}

// A workflow schedule runs as the workflow's own user, the one set on
// it, not the schedule's; a run-as user nobody holds skips the time, and
// the schedule moves on.
func TestLoop_TickWorkflowRunAs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	miguel := &model.User{Name: "miguel", Role: model.RoleRunner}
	assert.NoError(t, f.users.Create(ctx, miguel))
	wf, _ := f.workflow(t, "daily summary", "plan")
	wf.RunAsUserID = miguel.ID
	assert.NoError(t, f.workflows.Update(ctx, wf, []model.WorkflowStep{{Folder: "infra", Job: "plan"}}))
	gone, skipped := f.workflow(t, "orphan", "plan")
	gone.RunAsUserID = "nobody"
	assert.NoError(t, f.workflows.Update(ctx, gone, []model.WorkflowStep{{Folder: "infra", Job: "plan"}}))

	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)
	assert.Equal(t, []string{`daily summary as miguel with ""`}, f.flows.started())
	got, err := f.schedules.Get(ctx, skipped.ID)
	assert.NoError(t, err)
	assert.True(t, got.IsEnabled())
	assert.Equal(t, f.now.Add(time.Hour), got.NextRunAt.UTC(), "moved on")
}

// The loop disables a schedule whose workflow nobody holds, skips one
// whose workflow runs and one whose user may not run a step, and moves
// each on all the same.
func TestLoop_TickWorkflowSkipped(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	wf, gone := f.workflow(t, "gone", "plan")
	assert.NoError(t, f.workflows.Delete(ctx, wf.ID))
	f.loop.Catalogue.Find("infra").Job("slow").Allow = []string{"admin"}
	_, forbidden := f.workflow(t, "slow one", "plan", "slow")
	_, busy := f.workflow(t, "busy", "plan")
	f.flows.busy = true

	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	assert.Len(t, f.flows.started(), 0)
	got, err := f.schedules.Get(ctx, gone.ID)
	assert.NoError(t, err)
	assert.False(t, got.IsEnabled(), "the gone workflow's schedule is off")
	assert.Nil(t, got.LastRunAt)
	for _, schedule := range []*model.Schedule{forbidden, busy} {
		got, err := f.schedules.Get(ctx, schedule.ID)
		assert.NoError(t, err)
		assert.True(t, got.IsEnabled())
		assert.Equal(t, f.now.Add(time.Hour), got.NextRunAt.UTC(), "moved on")
	}

	// Without the runner to start it, a workflow schedule waits.
	f.flows.busy = false
	f.loop.Flows = nil
	_, waits := f.workflow(t, "waits", "plan")
	started, err = f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	got, err = f.schedules.Get(ctx, waits.ID)
	assert.NoError(t, err)
	assert.Nil(t, got.LastRunAt)
}

func TestLoop_Tick(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	schedule := f.schedule(t, "plan", "0 * * * *")
	user := f.ops

	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)

	got, err := f.schedules.Get(ctx, schedule.ID)
	assert.NoError(t, err)
	assert.Equal(t, f.now, got.LastRunAt.UTC())
	assert.Equal(t, f.now.Add(time.Hour), got.NextRunAt.UTC())

	runs, err := f.runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "plan"})
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, int64(1), runs[0].Scheduled)
	assert.Equal(t, "ops", runs[0].UserName)
	assert.Equal(t, user.ID, runs[0].UserID)

	// Not due again until its next time.
	started, err = f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	assert.Eventually(t, func() bool {
		active, _ := f.runs.Active(ctx)
		return len(active) == 0
	}, 5*time.Second, 10*time.Millisecond, "the first run finishes")
	f.now = f.now.Add(time.Hour)
	started, err = f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)
}

// A scheduled run gets the keys of the schedule's creator: the user the
// policy maps the key to gets it through the run's agent, another user's
// schedule gets no agent.
func TestLoop_TickKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	keys := &keystore.Local{Dir: t.TempDir(), Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db), Tmp: t.TempDir()}
	deploy, err := keys.Generate(ctx, keystore.Spec{Name: "deploy"})
	assert.NoError(t, err)
	assert.NoError(t, keys.Link(ctx, deploy.Name, nil, []string{"black"}))

	dir := t.TempDir()
	infra := &folder.Folder{ID: "infra", Dir: dir}
	script := `ssh-add -l > "$TASK_UI_USER.txt" 2>&1 || echo no-agent >> "$TASK_UI_USER.txt"`
	for _, id := range []string{"mapped", "unmapped"} {
		job := &folder.Job{Folder: "infra", ID: id, Command: []string{"sh", "-c", script}}
		job.Name = id
		infra.Jobs = append(infra.Jobs, job)
	}
	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{
		Runs:       runs,
		Recordings: &recording.Store{Root: t.TempDir()},
		Grace:      time.Second,
		Keys:       keys,
	})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	schedules := storage.NewScheduleStorage(db)
	users := storage.NewUserStorage(db)
	loop := scheduler.NewLoop(schedules, users, catalogue{"infra": infra}, manager)
	now := time.Date(2026, time.October, 17, 10, 0, 0, 0, time.UTC)
	loop.Now = func() time.Time { return now }

	due := now.Add(-time.Minute)
	for job, name := range map[string]string{"mapped": "black", "unmapped": "ops"} {
		user := &model.User{Name: name, Role: model.RoleRunner}
		assert.NoError(t, users.Create(ctx, user))
		schedule := &model.Schedule{Folder: "infra", Job: job, Expression: "0 * * * *", Enabled: 1, NextRunAt: &due, UserID: user.ID, UserName: name}
		assert.NoError(t, schedules.Create(ctx, schedule))
	}
	started, err := loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, started)
	assert.Eventually(t, func() bool {
		active, _ := runs.Active(ctx)
		return len(active) == 0
	}, 5*time.Second, 10*time.Millisecond, "both runs finish")

	mapped, err := os.ReadFile(filepath.Join(dir, "black.txt"))
	assert.NoError(t, err)
	assert.Contains(t, string(mapped), "deploy_ed25519 (ED25519)")
	unmapped, err := os.ReadFile(filepath.Join(dir, "ops.txt"))
	assert.NoError(t, err)
	assert.Contains(t, string(unmapped), "no-agent")
}

func TestLoop_TickBusy(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.schedule(t, "slow", "* * * * *")
	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)

	// The job is still running at the next minute: skipped, not queued.
	f.now = f.now.Add(time.Minute)
	started, err = f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	runs, err := f.runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "slow"})
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
}

func TestLoop_TickGone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	missing := f.schedule(t, "nothing", "@hourly")
	bad := f.schedule(t, "plan", "0 0 31 2 *")
	shell := f.schedule(t, "@shell", "@hourly")

	started, err := f.loop.Tick(ctx)
	assert.Error(t, err, "the schedule that never fires is reported")
	assert.Equal(t, 0, started)
	for _, id := range []string{missing.ID, bad.ID, shell.ID} {
		got, err := f.schedules.Get(ctx, id)
		assert.NoError(t, err)
		assert.False(t, got.IsEnabled(), "disabled rather than asked again")
	}
}

func TestLoop_Run(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.loop.Interval = 20 * time.Millisecond
	f.schedule(t, "plan", "* * * * *")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		f.loop.Run(ctx)
		close(done)
	}()
	assert.Eventually(t, func() bool {
		runs, _ := f.runs.List(context.Background(), storage.RunFilter{Folder: "infra", Job: "plan"})
		return len(runs) == 1
	}, 5*time.Second, 10*time.Millisecond)
	cancel()
	<-done
}

// A schedule runs as its user only while that user may run the job:
// the loop skips one the allow lists leave out, or one nobody holds.
func TestLoop_TickMayRun(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.loop.Catalogue.Find("infra").Job("slow").Allow = []string{"admin"}
	f.schedule(t, "slow", "* * * * *")
	ghost := f.schedule(t, "plan", "* * * * *")
	ghost.UserID, ghost.UserName = "gone", "gone"
	assert.NoError(t, f.schedules.Update(ctx, ghost))

	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	runs, err := f.runs.List(ctx, storage.RunFilter{Folder: "infra"})
	assert.NoError(t, err)
	assert.Len(t, runs, 0)
}

// A schedule of @host:* stays: the hosts permission decides whether it
// runs, and without the hosts to run on it starts nothing.
func TestLoop_HostSchedule(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, f.users.Create(ctx, admin))
	due := f.now.Add(-time.Minute)
	for _, user := range []*model.User{f.ops, admin} {
		schedule := &model.Schedule{Folder: model.Default, Job: model.HostAll, Args: `["uptime"]`, Expression: "@hourly", Enabled: 1, NextRunAt: &due, UserID: user.ID, UserName: user.Name}
		assert.NoError(t, f.schedules.Create(ctx, schedule))
	}
	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, started)
	schedules, err := f.schedules.List(ctx)
	assert.NoError(t, err)
	for _, schedule := range schedules {
		assert.Equal(t, int64(1), schedule.Enabled)
	}
}

// A schedule of a job that takes args starts its run with them.
func TestLoop_ScheduleArgs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	schedule := f.schedule(t, "greet", "@hourly")
	schedule.Args = `["there","now"]`
	assert.NoError(t, f.schedules.Update(ctx, schedule))
	started, err := f.loop.Tick(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, started)
	runs, err := f.runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "greet"})
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, []string{"there", "now"}, runs[0].ArgList())
}
