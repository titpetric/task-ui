package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// RunSummaryStorage keeps Claude's summary of each run that has one, by
// the run's ID.
type RunSummaryStorage struct {
	db DB
}

// NewRunSummaryStorage returns the storage over a pool.
func NewRunSummaryStorage(db DB) *RunSummaryStorage {
	return &RunSummaryStorage{db: db}
}

// Get returns a run's summary, or ErrNotFound.
func (s *RunSummaryStorage) Get(ctx context.Context, runID string) (*model.RunSummary, error) {
	summary, err := client(s.db).Get[model.RunSummary](ctx, "SELECT * FROM run_summary WHERE run_id = ?", runID)
	return summary, notFound(err)
}

// Set stores a run's summary, in place of one it had.
func (s *RunSummaryStorage) Set(ctx context.Context, summary *model.RunSummary) error {
	if summary.RunID == "" || summary.Summary == "" {
		return errors.New("a summary needs a run and some text")
	}
	if err := s.Delete(ctx, summary.RunID); err != nil {
		return err
	}
	summary.CreatedAt = stamp()
	return client(s.db).Insert(ctx, model.RunSummaryTable, summary)
}

// Delete removes a run's summary; a run without one is no error.
func (s *RunSummaryStorage) Delete(ctx context.Context, runID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM run_summary WHERE run_id = ?", runID)
}
