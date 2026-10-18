package scheduler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

type fixture struct {
	loop      *scheduler.Loop
	schedules *storage.ScheduleStorage
	runs      *storage.RunStorage
	users     *storage.UserStorage
	now       time.Time
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
	infra.Jobs = []*folder.Job{plan, slow}

	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{
		Runs:       runs,
		Recordings: &recording.Store{Root: t.TempDir()},
		Grace:      time.Second,
	})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	schedules := storage.NewScheduleStorage(db)
	users := storage.NewUserStorage(db)
	loop := scheduler.NewLoop(schedules, users, catalogue{"infra": infra}, manager)
	f := &fixture{loop: loop, schedules: schedules, runs: runs, users: users, now: time.Date(2026, time.October, 17, 10, 0, 0, 0, time.UTC)}
	loop.Now = func() time.Time { return f.now }
	return f
}

func (f *fixture) schedule(t *testing.T, job, expression string) *model.Schedule {
	t.Helper()
	due := f.now.Add(-time.Minute)
	schedule := &model.Schedule{Folder: "infra", Job: job, Expression: expression, Enabled: 1, NextRunAt: &due, UserName: "ops"}
	assert.NoError(t, f.schedules.Create(context.Background(), schedule))
	return schedule
}

func TestNewLoop(t *testing.T) {
	t.Parallel()
	var folders scheduler.Catalogue = catalogue{}
	loop := scheduler.NewLoop(nil, nil, folders, nil)
	assert.Equal(t, time.Minute, loop.Interval)
	assert.NotNil(t, loop.Now)
}

func TestLoop_Tick(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	schedule := f.schedule(t, "plan", "0 * * * *")
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, f.users.Create(ctx, user))
	schedule.UserID = user.ID
	assert.NoError(t, f.schedules.Update(ctx, schedule))

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

	started, err := f.loop.Tick(ctx)
	assert.Error(t, err, "the schedule that never fires is reported")
	assert.Equal(t, 0, started)
	for _, id := range []string{missing.ID, bad.ID} {
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
