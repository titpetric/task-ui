package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func jobs(id string, names ...string) *folder.Folder {
	f := &folder.Folder{ID: id}
	for _, name := range names {
		j := &folder.Job{Folder: id, ID: name}
		j.Name = name
		f.Jobs = append(f.Jobs, j)
	}
	return f
}

// The control centre has a column for each job name that repeats across
// folders, the most common first, up to n; the rest of a folder's jobs
// are its extra. A folder without jobs has no row.
func TestControl(t *testing.T) {
	t.Parallel()
	folders := []*folder.Folder{
		jobs("api", "build", "test", "push", "lint"),
		jobs("web", "build", "test", "serve"),
		jobs("docs", "build", "deploy"),
		jobs("empty"),
		jobs("infra", "plan", "apply", "test"),
	}

	c := view.NewControl(folders, 2)
	assert.Equal(t, []string{"build", "test"}, c.Columns)
	assert.Len(t, c.Rows, 4)

	api := c.Rows[0]
	assert.Equal(t, "api", api.Folder.ID)
	assert.Equal(t, "build", api.Cells[0].ID)
	assert.Equal(t, "test", api.Cells[1].ID)
	assert.Equal(t, []string{"push", "lint"}, names(api.Extra))

	docs := c.Rows[2]
	assert.True(t, docs.Cells[1] == nil, "docs has no test")
	assert.Equal(t, []string{"deploy"}, names(docs.Extra))

	// A name only one folder has is no column, however many columns
	// there may be.
	wide := view.NewControl(folders, 7)
	assert.Equal(t, []string{"build", "test"}, wide.Columns)
}

func names(js []*folder.Job) []string {
	out := []string{}
	for _, j := range js {
		out = append(out, j.ID)
	}
	return out
}
