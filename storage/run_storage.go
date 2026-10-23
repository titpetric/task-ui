package storage

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// RunStorage keeps the runs.
type RunStorage struct {
	db DB
}

// NewRunStorage returns the storage over a pool.
func NewRunStorage(db DB) *RunStorage {
	return &RunStorage{db: db}
}

// RunFilter narrows a listing. Zero values match everything. Folders
// keeps the runs of those folders, empty for all. Before and After name
// runs by ID: a page of the runs older than the one, or newer.
type RunFilter struct {
	Folder  string
	Folders []string
	// Job keeps the runs of a job; @host:* keeps those of every host
	// job, @host:<host>.
	Job string
	// Dir is the directory inside the folder of the job, with Job: ""
	// keeps the runs of the folder's own job, not those of the job of a
	// directory inside it with the same name.
	Dir    string
	State  model.State
	UserID string
	// UserName narrows to the runs a user started, by name, which a run
	// keeps after its user leaves.
	UserName string
	// BatchID keeps the runs started together, the host runs of one Run
	// on the inventory.
	BatchID string
	Before  string
	After   string
	Limit   int
}

// Create inserts a run as running, started now.
func (s *RunStorage) Create(ctx context.Context, run *model.Run) error {
	if run.Folder == "" || run.Job == "" {
		return errors.New("a run needs a folder and a job")
	}
	if run.ID == "" {
		run.ID = model.NewID()
	}
	if run.State == "" {
		run.State = model.StateRunning
	}
	now := stamp()
	if run.StartedAt == nil {
		run.StartedAt = now
	}
	run.CreatedAt = now
	run.UpdatedAt = now
	return client(s.db).Insert(ctx, model.RunTable, run)
}

// Get returns a run by ID.
func (s *RunStorage) Get(ctx context.Context, id string) (*model.Run, error) {
	run, err := client(s.db).Get[model.Run](ctx, "SELECT * FROM run WHERE id = ?", id)
	return run, notFound(err)
}

// List returns runs matching the filter, newest first.
func (s *RunStorage) List(ctx context.Context, filter RunFilter) ([]model.Run, error) {
	query, args := runQuery(filter)
	runs, err := client(s.db).Select[model.Run](ctx, query, args...)
	if filter.Before == "" && filter.After != "" {
		slices.Reverse(runs)
	}
	return runs, err
}

// cursor returns the condition that keeps the rows of a table older
// than the row before names, or newer than the row after names, and the
// order the query reads them in: newer rows read up from the cursor,
// oldest first, for the caller to reverse. A cursor on a missing row
// keeps none.
func cursor(table, before, after string) (string, []any, string) {
	switch {
	case before != "":
		return "(started_at, id) < (SELECT started_at, id FROM " + table + " WHERE id = ?)", []any{before}, "started_at DESC, id DESC"
	case after != "":
		return "(started_at, id) > (SELECT started_at, id FROM " + table + " WHERE id = ?)", []any{after}, "started_at ASC, id ASC"
	}
	return "", nil, "started_at DESC, id DESC"
}

// runQuery builds the listing from the filter.
func runQuery(filter RunFilter) (string, []any) {
	var where []string
	var args []any
	add := func(column, value string) {
		if value != "" {
			where = append(where, column+" = ?")
			args = append(args, value)
		}
	}
	add("folder", filter.Folder)
	if filter.Job == model.HostAll {
		where = append(where, "job LIKE ?")
		args = append(args, model.HostPrefix+"%")
	} else {
		add("job", filter.Job)
	}
	if filter.Job != "" {
		where = append(where, "dir = ?")
		args = append(args, filter.Dir)
	}
	add("state", filter.State)
	add("user_id", filter.UserID)
	add("user_name", filter.UserName)
	add("batch_id", filter.BatchID)
	if len(filter.Folders) > 0 {
		where = append(where, "folder IN (?"+strings.Repeat(", ?", len(filter.Folders)-1)+")")
		for _, folder := range filter.Folders {
			args = append(args, folder)
		}
	}
	page, pageArgs, order := cursor("run", filter.Before, filter.After)
	if page != "" {
		where = append(where, page)
		args = append(args, pageArgs...)
	}
	query := "SELECT * FROM run"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY " + order
	if filter.Limit > 0 {
		query += " LIMIT " + strconv.Itoa(filter.Limit)
	}
	return query, args
}

// Active returns the runs in progress, oldest first.
func (s *RunStorage) Active(ctx context.Context) ([]model.Run, error) {
	return client(s.db).Select[model.Run](ctx, "SELECT * FROM run WHERE state = ? ORDER BY started_at ASC, id ASC", model.StateRunning)
}

// Last returns the most recent run of a job.
func (s *RunStorage) Last(ctx context.Context, filter RunFilter) (*model.Run, error) {
	filter.Limit = 1
	runs, err := s.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, ErrNotFound
	}
	return &runs[0], nil
}

// Latest returns the latest finished run of every job of the folders,
// every folder for none, a run per folder, directory and job.
func (s *RunStorage) Latest(ctx context.Context, folders []string) ([]model.Run, error) {
	where := "state != ?"
	args := []any{model.StateRunning}
	if len(folders) > 0 {
		where += " AND folder IN (?" + strings.Repeat(", ?", len(folders)-1) + ")"
		for _, folder := range folders {
			args = append(args, folder)
		}
	}
	return client(s.db).Select[model.Run](ctx, `SELECT * FROM run WHERE id IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY folder, dir, job ORDER BY started_at DESC, id DESC) AS latest
			FROM run WHERE `+where+`
		) WHERE latest = 1
	) ORDER BY folder, dir, job`, args...)
}

// Update writes a run back, by ID.
func (s *RunStorage) Update(ctx context.Context, run *model.Run) error {
	run.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.RunTable, run, "id")
}

// Finish records how a run ended.
func (s *RunStorage) Finish(ctx context.Context, run *model.Run) error {
	if !model.Finished(run.State) {
		return errors.New("a run finishes in a final state, not " + run.State)
	}
	run.FinishedAt = stamp()
	return s.Update(ctx, run)
}

// Delete removes a run by ID.
func (s *RunStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM run WHERE id = ?", id)
}

// KillActive marks every run in progress as killed: what the instance
// does when it starts, for runs the previous process took down with it.
func (s *RunStorage) KillActive(ctx context.Context) (int64, error) {
	db := client(s.db)
	if err := db.Exec(ctx, "UPDATE run SET state = ?, finished_at = ?, updated_at = ? WHERE state = ?",
		model.StateKilled, stamp(), stamp(), model.StateRunning); err != nil {
		return 0, err
	}
	return db.RowsAffected(), nil
}

// FailQueued marks every queued run failed with the message: what the
// instance does when it starts, for the runs that waited in the queue of
// the previous process and never started.
func (s *RunStorage) FailQueued(ctx context.Context, message string) (int64, error) {
	db := client(s.db)
	now := stamp()
	if err := db.Exec(ctx, "UPDATE run SET state = ?, exit_code = 1, message = ?, finished_at = ?, updated_at = ? WHERE state = ?",
		model.StateFailed, message, now, now, model.StateQueued); err != nil {
		return 0, err
	}
	return db.RowsAffected(), nil
}

// Recorded returns the runs with a recording, oldest first, for the
// data retention rules: the recording files are theirs to measure.
func (s *RunStorage) Recorded(ctx context.Context) ([]model.Run, error) {
	return client(s.db).Select[model.Run](ctx, "SELECT * FROM run WHERE cast_path != '' ORDER BY started_at ASC, id ASC")
}

// Busy returns the IDs of the runs whose recordings stay whatever the
// retention rules say: the runs in progress, and every step run of a
// workflow run in progress, whose prompt reads their output.
func (s *RunStorage) Busy(ctx context.Context) (map[string]bool, error) {
	ids, err := client(s.db).Select[runID](ctx, `SELECT id FROM run WHERE state = ?
		UNION SELECT run_id AS id FROM workflow_run_step WHERE run_id != '' AND workflow_run_id IN (SELECT id FROM workflow_run WHERE state = ?)`, model.StateRunning, model.StateRunning)
	if err != nil {
		return nil, err
	}
	busy := map[string]bool{}
	for _, id := range ids {
		busy[id.ID] = true
	}
	return busy, nil
}

// runID is a run's ID, as Busy reads it.
type runID struct {
	ID string `db:"id"`
}
