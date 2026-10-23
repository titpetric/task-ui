// Package storage owns the SQL of task-ui.
//
// Queries run through github.com/titpetric/pdo, whose generic methods
// (Get[T], Select[T], Insert, Update) scan straight into the mig
// generated model types. A *pdo.PDO is request scoped and not safe for
// concurrent use, so every method allocates one over the shared *sqlx.DB
// pool. Callers outside this package issue no SQL; a question the app
// has gets a method here.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-bridget/mig/migrate"
	"github.com/jmoiron/sqlx"

	"github.com/titpetric/pdo"

	"github.com/titpetric/task-ui/schema"
)

// DB is the shared connection pool under every storage.
type DB = *sqlx.DB

// Project is the migration project name recorded in the migrations table.
const Project = "task-ui"

// ErrNotFound reports a row that is not there.
var ErrNotFound = errors.New("not found")

// Open opens a sqlite database. The driver is the program's to link:
// main imports modernc.org/sqlite for its side effect, as do the tests. An empty DSN or ":memory:" is a private
// in-memory database held on one connection, for tests;
// anything else is a file.
func Open(dsn string) (DB, error) {
	memory := dsn == "" || dsn == ":memory:"
	if memory {
		dsn = ":memory:"
	} else if !strings.Contains(dsn, "?") {
		dsn += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	}
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if memory {
		db.SetMaxOpenConns(1)
	}
	return db, nil
}

// Migrate applies the schema migrations to the database.
func Migrate(ctx context.Context, db DB) error {
	m, err := migrate.NewManager(db, schema.Migrations(), Project)
	if err != nil {
		return err
	}
	_, err = m.Apply(ctx)
	return err
}

// OpenAndMigrate opens the database and brings it up to date.
func OpenAndMigrate(ctx context.Context, dsn string) (DB, error) {
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := Migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// conn is a request scoped client, the one pdo type the rest of the
// package names: pdo stays imported in this file alone.
type conn = *pdo.PDO

// client returns a request scoped pdo client over the shared pool.
func client(db DB) conn {
	return pdo.New(db)
}

// notFound maps the driver's missing row to ErrNotFound.
func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// stamp returns the current time, for created_at and updated_at.
func stamp() *time.Time {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &now
}

// utc returns a time as the database keeps it, in UTC to the
// millisecond: a column compared as text against another time holds
// one zone, whatever the zone of the clock that made it. Nil stays nil.
func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	at := t.UTC().Truncate(time.Millisecond)
	return &at
}

// Vacuum rebuilds the database to free the space of deleted rows,
// while the server runs, and folds the write-ahead log into the file.
func Vacuum(ctx context.Context, db DB) error {
	if _, err := db.ExecContext(ctx, "VACUUM"); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}
