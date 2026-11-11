package phpfront_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// created reads the tables a module's own schema creates.
var created = regexp.MustCompile(`CREATE TABLE (?:IF NOT EXISTS )?([a-z_]+)`)

// TestFront_Modules is the front end stating its own shape: every directory
// under modules/ registers itself at startup, and the tables it says it
// bundles are the ones its schema/ creates.
//
// The registration runs the PHP, so a modinfo.php that names itself wrongly or
// claims a table it does not create fails here and not on an instance.
func TestFront_Modules(t *testing.T) {
	root, err := phpfront.Extract(t.TempDir())
	assert.NoError(t, err)

	front, err := phpfront.New(phpfront.NewOptions(root))
	assert.NoError(t, err)

	modules, err := front.Modules(context.Background())
	assert.NoError(t, err)

	registered := map[string]*phpfront.Module{}
	for _, module := range modules {
		registered[module.Plugin] = module
	}

	dirs, err := os.ReadDir(filepath.Join(root, phpfront.ModulesDir))
	assert.NoError(t, err)

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		t.Run(dir.Name(), func(t *testing.T) {
			module, ok := registered[dir.Name()]
			assert.True(t, ok, "the directory registered no module")
			if !ok {
				return
			}
			assert.Equal(t, "taskui/"+dir.Name(), module.Name())
			assert.NotEmpty(t, module.Title)
			assert.NotEmpty(t, module.Icon)
			assert.Equal(t, bundled(t, root, dir.Name()), module.Tables,
				"the tables the module declares are the ones its schema creates")
		})
	}
}

// bundled are the tables the migrations of a module create, in order.
func bundled(t *testing.T, root, module string) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(root, phpfront.ModulesDir, module, "schema", "*.up.sql"))
	assert.NoError(t, err)

	// Empty and not nil: a module that bundles no table writes setTables(array()).
	out := []string{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		assert.NoError(t, err)
		for _, match := range created.FindAllStringSubmatch(string(data), -1) {
			out = append(out, match[1])
		}
	}
	sort.Strings(out)
	return out
}
