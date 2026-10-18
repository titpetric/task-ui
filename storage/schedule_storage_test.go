package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestScheduleStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	schedule := &model.Schedule{Folder: "infra", Job: "plan", Expression: "0 * * * *", Enabled: 1, UserName: "ops"}
	assert.NoError(t, schedules.Create(ctx, schedule))
	assert.True(t, model.ValidID(schedule.ID))
	assert.Error(t, schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan"}), "no expression")
}

func TestScheduleStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	schedule := &model.Schedule{Folder: "infra", Job: "plan", Expression: "@hourly"}
	assert.NoError(t, schedules.Create(ctx, schedule))
	got, err := schedules.Get(ctx, schedule.ID)
	assert.NoError(t, err)
	assert.Equal(t, "@hourly", got.Expression)
	_, err = schedules.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestScheduleStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "web", Job: "build", Expression: "@daily"}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "infra", Job: "plan", Expression: "@daily"}))
	list, err := schedules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "infra", list[0].Folder)

	byFolder, err := schedules.ListByFolder(ctx, "web")
	assert.NoError(t, err)
	assert.Len(t, byFolder, 1)
}

func TestScheduleStorage_Due(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	now := time.Now().UTC()
	past, future := now.Add(-time.Minute), now.Add(time.Minute)
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "a", Job: "due", Expression: "* * * * *", Enabled: 1, NextRunAt: &past}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "a", Job: "later", Expression: "* * * * *", Enabled: 1, NextRunAt: &future}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "a", Job: "off", Expression: "* * * * *", Enabled: 0, NextRunAt: &past}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "a", Job: "never", Expression: "* * * * *", Enabled: 1}))

	due, err := schedules.Due(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, "due", due[0].Job)
	assert.True(t, due[0].Due(now))
}

func TestScheduleStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	schedule := &model.Schedule{Folder: "infra", Job: "plan", Expression: "@daily", Enabled: 1}
	assert.NoError(t, schedules.Create(ctx, schedule))
	next := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	schedule.Enabled = 0
	schedule.NextRunAt = &next
	assert.NoError(t, schedules.Update(ctx, schedule))
	got, err := schedules.Get(ctx, schedule.ID)
	assert.NoError(t, err)
	assert.False(t, got.IsEnabled())
	assert.Equal(t, next, got.NextRunAt.UTC())
}

func TestScheduleStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schedules := storage.NewScheduleStorage(newDB(t))
	schedule := &model.Schedule{Folder: "infra", Job: "plan", Expression: "@daily"}
	assert.NoError(t, schedules.Create(ctx, schedule))
	assert.NoError(t, schedules.Delete(ctx, schedule.ID))
	_, err := schedules.Get(ctx, schedule.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
