package phpscript

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Connection is the name a module opens the instance's database under:
// `new Database("task-ui")`. There is one, and every module shares it,
// because a module's tables are task-ui's tables in task-ui's database.
const Connection = "task-ui"

// Database is what resolves the connection a module's PHP opens. It
// answers one name with the instance's own pool and refuses the rest, so
// a module reaches the database it migrated into and no other.
//
// The pool is the instance's: sqlite on one connection. A module holding
// a transaction open holds the instance's only connection with it, which
// is why a module's work belongs in a statement and not in a long
// transaction.
type Database struct {
	DB *sqlx.DB
}

// NewDatabase returns the provider over a pool. A nil pool answers every
// name with an error, which is an instance whose modules store nothing.
func NewDatabase(db *sqlx.DB) *Database {
	return &Database{DB: db}
}

// Open returns the pool for the first name it knows.
func (d *Database) Open(_ context.Context, names ...string) (*sqlx.DB, error) {
	if d == nil || d.DB == nil {
		return nil, fmt.Errorf("no database on this instance: a module reads %q", Connection)
	}
	for _, name := range names {
		if name == Connection {
			return d.DB, nil
		}
	}
	return nil, fmt.Errorf("no connection named %v: a module reads %q", names, Connection)
}

// Connect is Open plus a ping, which is what the Database constructor
// calls: the ping proves the storage reachable before a script sees a
// client.
func (d *Database) Connect(ctx context.Context, names ...string) (*sqlx.DB, error) {
	db, err := d.Open(ctx, names...)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
