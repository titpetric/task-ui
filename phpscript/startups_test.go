package phpscript_test

import (
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// A module's startup files are every annotated PHP file but its
// modinfo.php, by path: that one runs first and on its own, because
// nothing may run before a module has registered.
func TestStartups(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, phpscript.ModInfo), startup)
	write(t, filepath.Join(root, "migrate.php"), "<?php\n// @startup\nTask::migrate(\"schema/\");\n")
	write(t, filepath.Join(root, "seed.php"), "<?php\n// @startup\n")
	write(t, filepath.Join(root, phpscript.Index), "<?php echo 'the page';")
	write(t, filepath.Join(root, "lib", "helpers.php"), "<?php\n// no annotation\n")
	write(t, filepath.Join(root, "vendor", "dep", "boot.php"), "<?php\n// @startup\n")
	write(t, filepath.Join(root, "notes.md"), "# @startup in prose is not a file to run\n")

	files, err := phpscript.Startups(root)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(root, "migrate.php"),
		filepath.Join(root, "seed.php"),
	}, files, "by path, without modinfo.php, and never out of a vendor tree")

	empty, err := phpscript.Startups(t.TempDir())
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}
