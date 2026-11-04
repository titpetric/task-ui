//go:build integration

package storage_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-bridget/mig/migrate"
	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// migrations is the schema of the instance, merged once for the binary: the
// owners are embedded trees and walking all four is work no test repeats.
//
// It comes from app and not from schema, because app is the one place that
// names the owners: see app/migrations.go.
var migrations = app.Migrations()

// newDB opens a private in-memory database with the schema applied.
func newDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", migrations)
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// closedDB is a migrated database that is closed, so every query
// through it fails with "sql: database is closed".
//
// It is how the error paths of this package are tested. A method here
// is a query and a check of what the driver said; the check is most of
// the branches in the package and none of them can be reached by a
// well-formed call against a working database. A closed pool reaches
// every one of them at once, which is why a storage test comes in two
// halves: what the rows do, and what a dead pool does.
func closedDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", migrations)
	assert.NoError(t, err)
	assert.NoError(t, db.Close())
	return db
}

// errClosed is what the driver says once the pool is closed, and what
// every method is asserted to pass back unchanged: this package wraps no
// driver error, so the string arrives as the driver wrote it.
const errClosed = "sql: database is closed"

// brokenDB is a migrated database with tables dropped. A closed pool
// fails before a transaction opens, so it never reaches the statements
// inside one; a missing table opens the transaction and fails in it,
// which is the other half of the error paths.
func brokenDB(t *testing.T, tables ...string) storage.DB {
	t.Helper()
	db := newDB(t)
	drop(t, db, tables...)
	return db
}

// drop removes tables from a database that already holds rows, for the
// methods that need one before the table goes.
func drop(t *testing.T, db storage.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := db.ExecContext(context.Background(), "DROP TABLE "+table)
		assert.NoError(t, err)
	}
}

// errNoTable is what the driver says for a statement against a table
// brokenDB dropped.
const errNoTable = "no such table"

// widenedDB is a migrated database with a column added to a table that
// the model type has no field for.
//
// It is the lever for the methods that write and then read the row back:
// a write names its columns and still succeeds, and the SELECT * that
// follows fails to scan. Nothing else reaches the second error of such a
// method, because the first one succeeding is what gets there.
func widenedDB(t *testing.T, table string) storage.DB {
	t.Helper()
	db := newDB(t)
	widen(t, db, table)
	return db
}

// widen adds the column to a table of a database that already holds
// rows, for the methods that need one to read back.
func widen(t *testing.T, db storage.DB, table string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "ALTER TABLE "+table+" ADD COLUMN unexpected TEXT NOT NULL DEFAULT ''")
	assert.NoError(t, err)
}

// errNoField is what sqlx says when a row holds a column the model type
// has no field for, which is what widenedDB arranges.
const errNoField = "missing destination name unexpected"

// refuse puts a trigger on a table that aborts every statement of one
// operation: INSERT, UPDATE or DELETE.
//
// It is the lever for the methods that read rows and then write them one
// at a time. The read has to succeed to reach the write, so nothing that
// breaks the table reaches it; a trigger leaves the read alone and fails
// the write the row after it.
func refuse(t *testing.T, db storage.DB, table, operation string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"CREATE TRIGGER no_"+operation+" BEFORE "+operation+" ON "+table+
			" BEGIN SELECT RAISE(ABORT, 'refused'); END")
	assert.NoError(t, err)
}

// errRefused is what the trigger refuse installed says.
const errRefused = "refused"

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
	assert.NoError(t, storage.Migrate(ctx, db, migrations))
	assert.NoError(t, storage.Migrate(ctx, db, migrations), "applying twice is fine")

	// As many rows as there are migrations embedded, counted rather than
	// written down: a migration added is a file added, and a number here
	// only says what it was when someone last looked. app's own test is
	// where the names are pinned.
	files, err := fs.Glob(migrations, "*.up.sql")
	assert.NoError(t, err)

	var count int
	assert.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM migrations WHERE project = ?", storage.Project))
	assert.Equal(t, len(files), count)
}

func TestOpenAndMigrate(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	var name string
	assert.NoError(t, db.GetContext(context.Background(), &name, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'run'"))
	assert.Equal(t, "run", name)

	_, err := storage.OpenAndMigrate(context.Background(), "file:/nonexistent/dir/x.db", migrations)
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
	statements, err := migrate.Statements(migrations,"schedule_utc.up.sql")
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
		statements, err := migrate.Statements(migrations,file)
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

// TestMigrate_stepKinds sets the kind of the steps of a workflow and of
// its runs from their fields, and moves a workflow's final prompt to a
// last prompt step, after its steps, in prompt_folder or the root;
// the workflow's prompt columns read empty after.
func TestMigrate_stepKinds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	for _, insert := range []string{
		`INSERT INTO workflow (id, name, prompt, prompt_folder) VALUES ('w1', 'porec', ' water today? ', ''), ('w2', 'deploy', 'summarize', 'malta/dockerfiles'), ('w3', 'plain', '', '')`,
		`INSERT INTO workflow_step (id, workflow_id, position, folder, job, prompt, workflow_ref, kind) VALUES
			('s1', 'w1', 0, '_', '@claude-prompt', 'fetch the weather', '', 'job'),
			('s2', 'w3', 0, '', '', '', 'w2', 'job'),
			('s3', 'w3', 1, 'infra', 'plan', '', '', 'job')`,
		`INSERT INTO workflow_run_step (id, workflow_run_id, position, folder, job, prompt, workflow_ref, kind) VALUES
			('r1', 'run', 0, '_', '@claude-prompt', 'fetch', '', 'job'),
			('r2', 'run', 1, '', '', '', 'w2', 'job')`,
	} {
		_, err := db.ExecContext(ctx, insert)
		assert.NoError(t, err)
	}
	statements, err := migrate.Statements(migrations,"workflow_step_with_kind.up.sql")
	assert.NoError(t, err)
	// The columns are there already; the rest of the migration runs
	// over the rows above.
	for _, statement := range statements {
		if strings.HasPrefix(strings.TrimSpace(statement), "ALTER TABLE") {
			continue
		}
		_, err := db.ExecContext(ctx, statement)
		assert.NoError(t, err)
	}
	type row struct {
		Workflow string `db:"workflow_id"`
		Position int64  `db:"position"`
		Folder   string `db:"folder"`
		Job      string `db:"job"`
		Kind     string `db:"kind"`
		Prompt   string `db:"prompt"`
	}
	var steps []row
	assert.NoError(t, db.SelectContext(ctx, &steps, "SELECT workflow_id, position, folder, job, kind, prompt FROM workflow_step ORDER BY workflow_id, position"))
	assert.Equal(t, []row{
		{"w1", 0, "_", "@claude-prompt", "prompt", "fetch the weather"},
		{"w1", 1, "_", "@claude-prompt", "prompt", "water today?"},
		{"w2", 0, "malta/dockerfiles", "@claude-prompt", "prompt", "summarize"},
		{"w3", 0, "", "", "workflow", ""},
		{"w3", 1, "infra", "plan", "job", ""},
	}, steps)
	var kinds []string
	assert.NoError(t, db.SelectContext(ctx, &kinds, "SELECT kind FROM workflow_run_step ORDER BY position"))
	assert.Equal(t, []string{"prompt", "workflow"}, kinds)
	var left int
	assert.NoError(t, db.GetContext(ctx, &left, "SELECT COUNT(*) FROM workflow WHERE prompt <> '' OR prompt_folder <> ''"))
	assert.Equal(t, 0, left)
}

// A database that cannot be migrated and a pool that is closed each
// hand back what they were told.
//
// Open itself has no error to test: sqlx.Open answers one only for a
// driver that is not registered, and the program links this one in.
func TestOpen_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	notADatabase := filepath.Join(t.TempDir(), "text.db")
	assert.NoError(t, os.WriteFile(notADatabase, []byte(strings.Repeat("not a database", 100)), 0o600))
	_, err := storage.OpenAndMigrate(ctx, notADatabase, migrations)
	assert.Error(t, err, "a file that is not a database")

	assert.ErrorContains(t, storage.Migrate(ctx, closedDB(t), migrations), errClosed)
	assert.ErrorContains(t, storage.Vacuum(ctx, closedDB(t)), errClosed)
}
