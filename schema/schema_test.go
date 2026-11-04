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
	assert.Len(t, files, 84)
	assert.Contains(t, files, "run.up.sql")
	assert.Contains(t, files, "workspace.up.sql")
	for _, name := range []string{"refinement_thread.up.sql", "refinement_comment.up.sql", "refinement_correction.up.sql"} {
		assert.Contains(t, files, name)
	}
	for _, name := range []string{"workflow.up.sql", "workflow_step.up.sql", "workflow_run.up.sql", "workflow_run_step.up.sql", "workflow_run_as.up.sql"} {
		assert.Contains(t, files, name)
	}
	data, err := fs.ReadFile(schema.Migrations(), "user.up.sql")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "CREATE TABLE IF NOT EXISTS user")

	// The two columns an ALTER added while the phpscript module owned
	// these tables are declared in the CREATE TABLE here. A database that
	// already holds the table fails that ALTER a second time, and SQLite
	// has no ADD COLUMN IF NOT EXISTS to soften it, so the column belongs
	// where every statement of these four files is conditional.
	assert.NotContains(t, files, "refinement_comment_collapse.up.sql")
	data, err = fs.ReadFile(schema.Migrations(), "refinement_comment.up.sql")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "kind TEXT NOT NULL DEFAULT ''")
	assert.Contains(t, string(data), "collapsed_into CHAR(26) NOT NULL DEFAULT ''")
}
