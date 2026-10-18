package schema_test

import (
	"io/fs"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/schema"
)

func TestMigrations(t *testing.T) {
	t.Parallel()
	files, err := fs.Glob(schema.Migrations(), "*.up.sql")
	assert.NoError(t, err)
	assert.Len(t, files, 7)
	assert.Contains(t, files, "run.up.sql")
	data, err := fs.ReadFile(schema.Migrations(), "user.up.sql")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "CREATE TABLE IF NOT EXISTS user")
}
