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
	assert.Len(t, files, 50)
	assert.Contains(t, files, "run.up.sql")
	assert.Contains(t, files, "workspace.up.sql")
	for _, name := range []string{"workflow.up.sql", "workflow_step.up.sql", "workflow_run.up.sql", "workflow_run_step.up.sql"} {
		assert.Contains(t, files, name)
	}
	data, err := fs.ReadFile(schema.Migrations(), "user.up.sql")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "CREATE TABLE IF NOT EXISTS user")
}
