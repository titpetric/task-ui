//go:build integration

package storage_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-bridget/mig/migrate"
	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/schema"
	"github.com/titpetric/task-ui/storage"
)

// newDB opens a private in-memory database with the schema applied.
func newDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpen(t *testing.T) {
	t.Parallel()
	db, err := storage.Open("")
	assert.NoError(t, err)
	var pool storage.DB = db
	assert.NoError(t, pool.Ping())
	assert.Equal(t, 1, db.Stats().MaxOpenConnections)
	db.Close()

	file, err := storage.Open("file:" + filepath.Join(t.TempDir(), "x.db"))
	assert.NoError(t, err)
	assert.NoError(t, file.Ping())
	file.Close()
}

func TestMigrate(t *testing.T) {
	t.Parallel()
	db, err := storage.Open(":memory:")
	assert.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	assert.NoError(t, storage.Migrate(ctx, db))
	assert.NoError(t, storage.Migrate(ctx, db), "applying twice is fine")

	var count int
	assert.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM migrations WHERE project = ?", storage.Project))
	assert.Equal(t, 60, count)
}

func TestOpenAndMigrate(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	var name string
	assert.NoError(t, db.GetContext(context.Background(), &name, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'run'"))
	assert.Equal(t, "run", name)

	_, err := storage.OpenAndMigrate(context.Background(), "file:/nonexistent/dir/x.db")
	assert.Error(t, err)
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()
	assert.Error(t, storage.ErrNotFound)
	assert.Equal(t, "not found", storage.ErrNotFound.Error())
}

// Vacuum works on a live database.
func TestVacuum(t *testing.T) {
	t.Parallel()
	assert.NoError(t, storage.Vacuum(context.Background(), newDB(t)))
}

// TestMigrate_scheduleUTC rewrites the times of schedules written with
// the local offset, as Go's time text has them, to the same instants
// in UTC.
func TestMigrate_scheduleUTC(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	_, err := db.ExecContext(ctx, `INSERT INTO schedule (id, folder, job, expression, next_run_at, last_run_at) VALUES
		('a', '_', 'weather', '0 7 * * *', '2026-10-23 07:00:00 +0200 CEST', '2026-10-22 09:00:28.172944004 +0200 CEST m=+26340.112136230'),
		('b', '_', 'west', '@daily', '2026-01-10 00:30:00 -0330 NST', NULL),
		('c', '_', 'utc', '@daily', '2026-10-23 05:00:00.5 +0000 UTC', NULL)`)
	assert.NoError(t, err)
	statements, err := migrate.Statements(schema.Migrations(), "schedule_utc.up.sql")
	assert.NoError(t, err)
	for _, statement := range statements {
		_, err := db.ExecContext(ctx, statement)
		assert.NoError(t, err)
	}
	type times struct {
		Next string  `db:"next_run_at"`
		Last *string `db:"last_run_at"`
	}
	var got []times
	assert.NoError(t, db.SelectContext(ctx, &got, "SELECT CAST(next_run_at AS TEXT) AS next_run_at, CAST(last_run_at AS TEXT) AS last_run_at FROM schedule ORDER BY id"))
	assert.Len(t, got, 3)
	assert.Equal(t, "2026-10-23 05:00:00 +0000 UTC", got[0].Next)
	assert.Equal(t, "2026-10-22 07:00:28 +0000 UTC", *got[0].Last)
	assert.Equal(t, "2026-01-10 04:00:00 +0000 UTC", got[1].Next)
	assert.Nil(t, got[1].Last)
	assert.Equal(t, "2026-10-23 05:00:00.5 +0000 UTC", got[2].Next, "a time in UTC stays")
}

// TestMigrate_hostJobs moves the runs, schedules and steps of the bare
// @host to the job of their target, @host:<host> or @host:*; the runs
// keep the target, the schedules and steps drop it.
func TestMigrate_hostJobs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	for _, insert := range []string{
		`INSERT INTO run (id, folder, job, target, state) VALUES ('r1', '_', '@host', 'dev1', 'ok'), ('r2', '_', 'plan', '', 'ok')`,
		`INSERT INTO schedule (id, folder, job, expression, target, args) VALUES ('s1', '_', '@host', '@daily', '*', '["uptime"]')`,
		`INSERT INTO workflow_step (id, workflow_id, position, folder, job, target, args) VALUES ('w1', 'wf', 0, '_', '@host', 'dev2', '["uptime"]')`,
	} {
		_, err := db.ExecContext(ctx, insert)
		assert.NoError(t, err)
	}
	for _, file := range []string{"run_target_host_job.up.sql", "schedule_target_host_job.up.sql", "workflow_step_target_host_job.up.sql"} {
		statements, err := migrate.Statements(schema.Migrations(), file)
		assert.NoError(t, err)
		for _, statement := range statements {
			_, err := db.ExecContext(ctx, statement)
			assert.NoError(t, err)
		}
	}
	type row struct {
		Job    string `db:"job"`
		Target string `db:"target"`
	}
	var runs []row
	assert.NoError(t, db.SelectContext(ctx, &runs, "SELECT job, target FROM run ORDER BY id"))
	assert.Equal(t, []row{{"@host:dev1", "dev1"}, {"plan", ""}}, runs)
	var schedule, step row
	assert.NoError(t, db.GetContext(ctx, &schedule, "SELECT job, target FROM schedule"))
	assert.Equal(t, row{"@host:*", ""}, schedule)
	assert.NoError(t, db.GetContext(ctx, &step, "SELECT job, target FROM workflow_step"))
	assert.Equal(t, row{"@host:dev2", ""}, step)
}
