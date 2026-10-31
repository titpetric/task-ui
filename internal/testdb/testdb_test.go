package testdb_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testdb"
)

// A test's database is migrated, and it is its own: a row one test writes
// is not there for the next, because the template is copied and not
// shared.
func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := testdb.New(t)

	var tables int
	assert.NoError(t, db.GetContext(ctx, &tables,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'user'"))
	assert.Equal(t, 1, tables, "the schema is applied")

	_, err := db.ExecContext(ctx, "INSERT INTO user (id, name) VALUES ('01TESTUSER0000000000000001', 'ana')")
	assert.NoError(t, err)

	var users int
	assert.NoError(t, db.GetContext(ctx, &users, "SELECT COUNT(*) FROM user"))
	assert.Equal(t, 1, users)

	assert.NoError(t, testdb.New(t).GetContext(ctx, &users, "SELECT COUNT(*) FROM user"))
	assert.Equal(t, 0, users, "a second database holds nothing the first wrote")
}
