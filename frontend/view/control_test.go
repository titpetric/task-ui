package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// jobs makes a folder of jobs: "name" or "kind:name".
func jobs(id string, ids ...string) *folder.Folder {
	f := &folder.Folder{ID: id}
	for _, jobID := range ids {
		j := &folder.Job{Folder: id, ID: jobID}
		j.Name = jobID
		if kind, name, ok := cutKind(jobID); ok {
			j.Kind, j.Name = kind, name
		}
		f.Jobs = append(f.Jobs, j)
	}
	return f
}

func cutKind(id string) (string, string, bool) {
	for i := range id {
		if id[i] == ':' {
			return id[:i], id[i+1:], true
		}
	}
	return "", "", false
}

// Every job name that occurs more than once, across folders or within
// one, gets a column, the most common first; a cell holds all of the
// folder's jobs of that name. Other holds the names that occur once. A
// folder without jobs has no row.
func TestControl(t *testing.T) {
	t.Parallel()
	folders := []*folder.Folder{
		jobs("api", "build", "test", "push", "lint"),
		jobs("web", "build", "test", "serve"),
		jobs("docs", "build", "deploy"),
		jobs("empty"),
		jobs("infra", "plan", "apply", "test"),
		jobs("tools", "make:release", "script:release", "make:lint"),
	}

	c := view.NewControl(folders)
	assert.Equal(t, []string{"build", "test", "lint", "release"}, c.Columns)
	assert.Len(t, c.Rows, 5)

	api := c.Rows[0]
	assert.Equal(t, "api", api.Folder.ID)
	assert.Equal(t, []string{"build"}, names(api.Cells[0]))
	assert.Equal(t, []string{"lint"}, names(api.Cells[2]))
	assert.Equal(t, []string{"push"}, names(api.Extra))

	docs := c.Rows[2]
	assert.Len(t, docs.Cells[1], 0) // docs has no test
	assert.Equal(t, []string{"deploy"}, names(docs.Extra))

	// Two of a name in one folder share its cell; nothing repeats under
	// other.
	tools := c.Rows[4]
	assert.Equal(t, []string{"make:release", "script:release"}, names(tools.Cells[3]))
	assert.Equal(t, []string{"make:lint"}, names(tools.Cells[2]))
	assert.Len(t, tools.Extra, 0)
}

func names(js []*folder.Job) []string {
	out := []string{}
	for _, j := range js {
		out = append(out, j.ID)
	}
	return out
}
