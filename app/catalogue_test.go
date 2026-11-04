package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/layout"
)

// projects writes n projects with a job each in a workspace's src.
func projects(t *testing.T, root, workspace string, n int) {
	t.Helper()
	for i := range n {
		dir := filepath.Join(layout.SrcDir(root, workspace), "p"+strconv.Itoa(i))
		assert.NoError(t, os.MkdirAll(dir, 0o755))
		body := "tasks:\n  build:\n    desc: Build\n"
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(body), 0o644))
	}
}

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
	assert.Equal(t, "internal", steps.Find("internal").ID)
	assert.Equal(t, 2, len(c.Every()), "cached")
}

// A listing of a workspace's projects takes the same walks whatever the
// number of projects, and Find answers out of the catalogue it already
// holds.
//
// It was the other way around, and it was the latency of every page. A
// walk reads the project and workspace rows and goes into every checkout
// under the root; Steps().Find took one to answer a single ID. Asking it
// per directory walked the root once per project, and the sidebar does
// ask, because a workflow that runs over a workspace reads its projects
// to say whether the reader may start it. The ported front end asks for
// the job of each step one Host::job at a time, which is the same walk
// again. The page took three and a half seconds on an instance whose
// database is three megabytes.
func TestCatalogueWalks(t *testing.T) {
	t.Parallel()
	walks := func(n int) int64 {
		root := t.TempDir()
		projects(t, root, "ws", n)
		c := app.NewCatalogue(root, true, nil)
		assert.Len(t, c.WorkspaceProjects(context.Background(), "ws"), n)
		return c.Walks()
	}
	// Two: the catalogue as the pages see it, and the Every beside it
	// that holds the folders a section or a setting hides.
	assert.Equal(t, int64(2), walks(2))
	assert.Equal(t, int64(2), walks(20))

	root := t.TempDir()
	projects(t, root, "", 20)
	c := app.NewCatalogue(root, true, nil)
	for i := range 20 {
		assert.True(t, c.Find("p"+strconv.Itoa(i)) != nil)
	}
	assert.Equal(t, int64(1), c.Walks())
}
