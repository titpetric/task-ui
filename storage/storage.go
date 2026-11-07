// Package storage owns the SQL of task-ui.
//
// Queries run through github.com/titpetric/pdo, whose generic methods
// (Get[T], Select[T], Insert, Update) scan straight into the mig
// generated model types. A *pdo.PDO is request scoped and not safe for
// concurrent use, so every method allocates one over the shared *sqlx.DB
// pool. Every SQL statement task-ui runs is in this package; a question
// the app has gets a method here. rules/sql.yml refuses one written in
// another package.
//
// Three shapes recur. The reason for each is here, so no method repeats it.
//
// A method that is one statement takes no transaction. The statement
// either happens whole or not at all, and a count comes from the
// driver's RowsAffected. Keep and MoveFolder are written that way; inTx
// is for the methods that genuinely write across tables.
//
// A method that names its columns is one task-ui migrate-layout calls. It
// runs against a database from before the migration it is part of, whose
// tables hold other columns than this package's models do, and SELECT *
// would scan them.
//
// No method wraps a driver error. A caller comparing against sql.ErrNoRows
// or reading a message gets the driver's own; ErrNotFound is the one
// translation, and notFound makes it.
//
// Migrate takes the migrations as a parameter because they live with the
// code that owns each table, and some of those owners import this package.
// app.Migrations merges them, and app is the only place naming the owners.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"strings"
	"time"

	"github.com/go-bridget/mig/migrate"
	"github.com/jmoiron/sqlx"

	"github.com/titpetric/pdo"
)

// DB is the shared connection pool under every storage. The program
// links the driver: main and the tests import modernc.org/sqlite for
// its side effect.
type DB = *sqlx.DB

// Project is the migration project name recorded in the migrations table.
const Project = "task-ui"

// ErrNotFound reports a row that is not there.
var ErrNotFound = errors.New("not found")

// Open opens a sqlite database: a file, or for an empty DSN or ":memory:" a private one in memory on one connection.
func Open(dsn string) (DB, error) {
	memory := dsn == "" || dsn == ":memory:"
	if memory {
		dsn = ":memory:"
	} else if !strings.Contains(dsn, "?") {
		dsn += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	}
	// sqlx.Open answers an error only for a driver that is not
	// registered, and the program links this one in.
	db, err := sqlx.Open("sqlite", dsn)
	if memory && err == nil {
		db.SetMaxOpenConns(1)
	}
	return db, err
}

// Migrate applies the migrations of fsys to the database. The package doc says why fsys is a parameter.
func Migrate(ctx context.Context, db DB, fsys fs.FS) error {
	m, err := migrate.NewManager(db, fsys, Project)
	if err != nil {
		return err
	}
	_, err = m.Apply(ctx)
	return err
}

// OpenAndMigrate opens the database and brings it up to date.
func OpenAndMigrate(ctx context.Context, dsn string, fsys fs.FS) (DB, error) {
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := Migrate(ctx, db, fsys); err != nil {
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

// notIn is the clause that excludes a list of values from a statement,
// " <keyword> <column> NOT IN (?, ?, ?)", and nothing at all for an
// empty list: a delete of everything but nothing is a delete of
// everything. The keyword is WHERE or AND, whichever the statement is at.
func notIn(keyword, column string, n int) string {
	if n == 0 {
		return ""
	}
	return " " + keyword + " " + column + " NOT IN (?" + strings.Repeat(", ?", n-1) + ")"
}

// anyList is a list of strings as query arguments.
func anyList(values []string) []any {
	args := make([]any, 0, len(values))
	for _, value := range values {
		args = append(args, value)
	}
	return args
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

// utc returns a time as the database keeps it, in UTC: a column
// compared as text against another time holds one zone, from any clock
// that made it. Nil stays nil.
//
// It keeps the precision the time came with. stamp() rounds the clock it reads
// to the millisecond, which is resolution enough for a row nobody else
// is racing; a time from a caller is a moment it measured, and two
// of them in one millisecond are two moments. Rounding them together
// makes a tie, and a list ordered by started_at then falls back to the
// id for an order the caller did not mean.
func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	at := t.UTC()
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
