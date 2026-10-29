package phpscript_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestMigrator(t *testing.T) {
	t.Parallel()
	var migrator phpscript.Migrator = &fakeMigrator{}
	m := &phpscript.Module{Vendor: "titpetric", Plugin: "refinement", Root: "/x"}

	assert.NoError(t, migrator.Migrate(t.Context(), m, "schema"))
	assert.Equal(t, []string{"titpetric/refinement:schema"}, migrator.(*fakeMigrator).dirs)
}
