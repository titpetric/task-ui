package storage

import (
	"context"
	"errors"
	"time"

	"github.com/titpetric/task-ui/model"
)

// ScheduleStorage keeps the schedules.
type ScheduleStorage struct {
	db DB
}

// NewScheduleStorage returns the storage over a pool.
func NewScheduleStorage(db DB) *ScheduleStorage {
	return &ScheduleStorage{db: db}
}

// Create inserts a schedule. NextRunAt is the caller's to set: the
// scheduler knows the expression, this package does not.
func (s *ScheduleStorage) Create(ctx context.Context, schedule *model.Schedule) error {
	if schedule.Folder == "" || schedule.Job == "" || schedule.Expression == "" {
		return errors.New("a schedule needs a folder, a job and an expression")
	}
	if schedule.ID == "" {
		schedule.ID = model.NewID()
	}
	if schedule.Args == "" {
		schedule.Args = "[]"
	}
	schedule.CreatedAt = stamp()
	schedule.UpdatedAt = schedule.CreatedAt
	return client(s.db).Insert(ctx, model.ScheduleTable, schedule)
}

// Get returns a schedule by ID.
func (s *ScheduleStorage) Get(ctx context.Context, id string) (*model.Schedule, error) {
	schedule, err := client(s.db).Get[model.Schedule](ctx, "SELECT * FROM schedule WHERE id = ?", id)
	return schedule, notFound(err)
}

// List returns every schedule, by folder and job.
func (s *ScheduleStorage) List(ctx context.Context) ([]model.Schedule, error) {
	return client(s.db).Select[model.Schedule](ctx, "SELECT * FROM schedule ORDER BY folder, job, id")
}

// ListByFolder returns the schedules of a folder.
func (s *ScheduleStorage) ListByFolder(ctx context.Context, folder string) ([]model.Schedule, error) {
	return client(s.db).Select[model.Schedule](ctx, "SELECT * FROM schedule WHERE folder = ? ORDER BY job, id", folder)
}

// Due returns the enabled schedules whose next run is not after now.
func (s *ScheduleStorage) Due(ctx context.Context, now time.Time) ([]model.Schedule, error) {
	return client(s.db).Select[model.Schedule](ctx, "SELECT * FROM schedule WHERE enabled = 1 AND next_run_at IS NOT NULL AND next_run_at <= ? ORDER BY next_run_at", now.UTC())
}

// Update writes a schedule back, by ID.
func (s *ScheduleStorage) Update(ctx context.Context, schedule *model.Schedule) error {
	schedule.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.ScheduleTable, schedule, "id")
}

// Delete removes a schedule by ID.
func (s *ScheduleStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM schedule WHERE id = ?", id)
}
