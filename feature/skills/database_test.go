package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/feature/skills"
	"github.com/titpetric/task-ui/internal/assert"
)

// ATKINS_DB names the home's history first and the instance's mirror second.
func TestAtkinsDB(t *testing.T) {
	t.Parallel()
	sep := string(os.PathListSeparator)
	assert.Equal(t, filepath.FromSlash("/data/malta/home/.atkins/atkins.db")+sep+filepath.FromSlash("/root/.atkins/atkins.db"),
		skills.AtkinsDB(filepath.FromSlash("/data/malta/home"), filepath.FromSlash("/root")))
	assert.Equal(t, filepath.FromSlash("/data/home/.atkins/atkins.db"),
		skills.AtkinsDB(filepath.FromSlash("/data/home"), ""))
}
