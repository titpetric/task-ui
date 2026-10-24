package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestCatalogueSteps(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for path, body := range map[string]string{
		"proj/Taskfile.yml":  "tasks:\n  build:\n    desc: Build\n",
		"proj/sub/Makefile":  "test:\n\techo test\n",
		"proj/docs/index.md": "nothing to run",
	} {
		path = filepath.Join(root, "src", filepath.FromSlash(path))
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}
	c := app.NewCatalogue(root, true, nil)

	// Below a pipeline folder the sidebar lists pipelines alone.
	assert.True(t, c.Find("proj/sub") == nil)
	steps := c.Steps()
	ids := []string{}
	for _, f := range steps.List() {
		ids = append(ids, f.ID)
	}
	assert.Equal(t, []string{"proj", "proj/sub"}, ids)
	sub := steps.Find("proj/sub")
	assert.True(t, sub != nil && sub.Job("test") != nil)
	assert.Equal(t, "proj", steps.Find("proj").ID)
	assert.True(t, steps.Find("proj/docs") == nil)
	assert.Equal(t, "_", steps.Find("_").ID)
	assert.Equal(t, 2, len(c.Every()), "cached")
}
