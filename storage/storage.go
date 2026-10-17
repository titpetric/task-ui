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
	_ "modernc.org/sqlite" // the sqlite driver, registered as "sqlite"

	"github.com/titpetric/task-ui/schema"
)

// Project is the migration project name recorded in the migrations table.
const Project = "task-ui"

// ErrNotFound is returned when a row is not there.
var ErrNotFound = errors.New("not found")

// Open opens a sqlite database. An empty DSN or ":memory:" is a private
// in-memory database held on one connection, which is what tests use;
// anything else is a file.
func Open(dsn string) (*sqlx.DB, error) {
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
func Migrate(ctx context.Context, db *sqlx.DB) error {
	m, err := migrate.NewManager(db, schema.Migrations(), Project)
	if err != nil {
		return err
	}
	_, err = m.Apply(ctx)
	return err
}

// OpenAndMigrate opens the database and brings it up to date.
func OpenAndMigrate(ctx context.Context, dsn string) (*sqlx.DB, error) {
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

// client returns a request scoped pdo client over the shared pool.
func client(db *sqlx.DB) *pdo.PDO {
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
