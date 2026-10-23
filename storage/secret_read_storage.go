package storage

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/titpetric/task-ui/model"
)

// SecretReadKeep is how many reads of one secret the log keeps: the
// newest, the older ones pruned as runs add more.
const SecretReadKeep = 1000

// SecretReadStorage keeps which run got which secret in its
// environment: one row per secret per run, the newest SecretReadKeep of
// each secret.
type SecretReadStorage struct {
	db DB
	// Keep is how many reads of a secret Record leaves, SecretReadKeep
	// unless set.
	Keep int
}

// NewSecretReadStorage returns the storage over a pool.
func NewSecretReadStorage(db DB) *SecretReadStorage {
	return &SecretReadStorage{db: db, Keep: SecretReadKeep}
}

// Record adds the reads of one run, then prunes each secret's log to the
// newest Keep.
func (s *SecretReadStorage) Record(ctx context.Context, reads []model.SecretRead) error {
	if len(reads) == 0 {
		return nil
	}
	return inTx(ctx, s.db, func(c conn) error {
		now := stamp()
		for i := range reads {
			read := &reads[i]
			if read.SecretID == "" || read.RunID == "" {
				return errors.New("a secret read needs a secret and a run")
			}
			read.ID = model.NewID()
			read.CreatedAt = now
			if err := c.Insert(ctx, model.SecretReadTable, read); err != nil {
				return err
			}
			if err := c.Exec(ctx, "DELETE FROM secret_read WHERE secret_id = ? AND rowid NOT IN (SELECT rowid FROM secret_read WHERE secret_id = ? ORDER BY rowid DESC LIMIT ?)", read.SecretID, read.SecretID, s.keep()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SecretReadStorage) keep() int {
	if s.Keep > 0 {
		return s.Keep
	}
	return SecretReadKeep
}

// List returns a secret's reads, newest first, limit at most: before
// and after name the read the page starts past, older or newer, as the
// runs' pager does.
func (s *SecretReadStorage) List(ctx context.Context, secretID, before, after string, limit int) ([]model.SecretRead, error) {
	query := "SELECT * FROM secret_read WHERE secret_id = ?"
	args := []any{secretID}
	// Reads of one run share a millisecond, so the order is the row's,
	// the order Record added them in.
	order := " ORDER BY rowid DESC"
	switch {
	case before != "":
		query += " AND rowid < (SELECT rowid FROM secret_read WHERE id = ?)"
		args = append(args, before)
	case after != "":
		query += " AND rowid > (SELECT rowid FROM secret_read WHERE id = ?)"
		args = append(args, after)
		order = " ORDER BY rowid ASC"
	}
	query += order
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	reads, err := client(s.db).Select[model.SecretRead](ctx, query, args...)
	if before == "" && after != "" {
		slices.Reverse(reads)
	}
	return reads, err
}

// ForRun returns the reads of one run, by secret name.
func (s *SecretReadStorage) ForRun(ctx context.Context, runID string) ([]model.SecretRead, error) {
	return client(s.db).Select[model.SecretRead](ctx, "SELECT * FROM secret_read WHERE run_id = ? ORDER BY name", runID)
}

// secretReadCount is a secret's number of reads, as Counts reads it.
type secretReadCount struct {
	SecretID string `db:"secret_id"`
	Count    int    `db:"n"`
}

// Counts returns how many reads the log holds of each secret, by secret
// ID; a secret no run read is missing.
func (s *SecretReadStorage) Counts(ctx context.Context) (map[string]int, error) {
	rows, err := client(s.db).Select[secretReadCount](ctx, "SELECT secret_id, COUNT(*) AS n FROM secret_read GROUP BY secret_id")
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.SecretID] = row.Count
	}
	return counts, nil
}

// Last returns the newest read of each secret, by secret ID.
func (s *SecretReadStorage) Last(ctx context.Context) (map[string]model.SecretRead, error) {
	rows, err := client(s.db).Select[model.SecretRead](ctx, "SELECT * FROM secret_read WHERE rowid IN (SELECT MAX(rowid) FROM secret_read GROUP BY secret_id)")
	if err != nil {
		return nil, err
	}
	last := map[string]model.SecretRead{}
	for _, row := range rows {
		last[row.SecretID] = row
	}
	return last, nil
}

// secretReadUsage is the log's size, as Usage reads it.
type secretReadUsage struct {
	Rows  int   `db:"n"`
	Bytes int64 `db:"bytes"`
}

// Usage returns how many reads the log holds and about how many bytes
// of text they take, for a page on what the instance keeps.
func (s *SecretReadStorage) Usage(ctx context.Context) (int, int64, error) {
	usage, err := client(s.db).Get[secretReadUsage](ctx, "SELECT COUNT(*) AS n, COALESCE(SUM(LENGTH(id) + LENGTH(secret_id) + LENGTH(name) + LENGTH(run_id) + LENGTH(folder) + LENGTH(dir) + LENGTH(job) + LENGTH(user_name) + 24), 0) AS bytes FROM secret_read")
	if err != nil {
		return 0, 0, err
	}
	return usage.Rows, usage.Bytes, nil
}

// Prune removes the reads older than the time, for a retention
// setting; the per-secret cap holds regardless.
func (s *SecretReadStorage) Prune(ctx context.Context, before time.Time) error {
	return client(s.db).Exec(ctx, "DELETE FROM secret_read WHERE id < ?", model.NewIDAt(before)[:10])
}
