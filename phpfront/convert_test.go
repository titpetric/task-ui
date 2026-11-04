package phpfront

import (
	"testing"

	"github.com/titpetric/phpscript/model"

	"github.com/titpetric/task-ui/internal/assert"
)

// A Go struct crosses into the runtime as a host object whose fields no key
// reaches, so the boundary converts to an array keyed by the names host.go
// declares. These hold the keys; bind_test holds that a script indexes by them.

// at is the nth row of an array, as an array.
func at(t *testing.T, a *model.Array, n int) *model.Array {
	t.Helper()

	var rows []*model.Array
	a.Range(func(_, value any) bool {
		row, ok := value.(*model.Array)
		assert.True(t, ok, "every row is an array")
		rows = append(rows, row)
		return true
	})
	assert.True(t, n < len(rows), "row out of range")

	return rows[n]
}

// text is a key of an array as a string.
func text(t *testing.T, a *model.Array, key string) string {
	t.Helper()

	value, ok := a.Get(key)
	assert.True(t, ok, key)
	out, ok := value.(string)
	assert.True(t, ok, key+" is a string")

	return out
}

// flag is a key of an array as a bool.
func flag(t *testing.T, a *model.Array, key string) bool {
	t.Helper()

	value, ok := a.Get(key)
	assert.True(t, ok, key)
	out, ok := value.(bool)
	assert.True(t, ok, key+" is a bool")

	return out
}

// list is a key of an array as an array.
func list(t *testing.T, a *model.Array, key string) *model.Array {
	t.Helper()

	value, ok := a.Get(key)
	assert.True(t, ok, key)
	out, ok := value.(*model.Array)
	assert.True(t, ok, key+" is an array")

	return out
}

func TestFolders(t *testing.T) {
	t.Parallel()

	out := folders([]Folder{{
		ID:    "infra",
		Title: "Infra",
		Dir:   "/src/infra",
		Allow: []string{"ops"},
		Jobs:  []Job{{ID: "build", Name: "build", Allow: []string{"ann"}}, {ID: "task:test", Name: "test"}},
	}})
	assert.Equal(t, 1, out.Len())

	row := at(t, out, 0)
	assert.Equal(t, "infra", text(t, row, "id"))
	assert.Equal(t, "Infra", text(t, row, "title"))
	assert.Equal(t, "/src/infra", text(t, row, "dir"))
	assert.Equal(t, 1, list(t, row, "allow").Len())

	// A job of a folder that follows more than one convention is named
	// kind:name and called by its name alone, so the two are separate keys:
	// a schedule is written with the ID and the picker shows the name.
	jobs := list(t, row, "jobs")
	assert.Equal(t, 2, jobs.Len())
	assert.Equal(t, "task:test", text(t, at(t, jobs, 1), "id"))
	assert.Equal(t, "test", text(t, at(t, jobs, 1), "name"))
	assert.Equal(t, 1, list(t, at(t, jobs, 0), "allow").Len())
}

// The job of a run answers the whole gate: whether it is there, the kind a
// permission gates, and both allow lists. A job that is not there answers
// every key empty, which is the nil pointer the Go page gates on.
func TestJob(t *testing.T) {
	t.Parallel()

	out := job(Job{Found: true, ID: "@host:db1", Name: "host db1", Host: true, FolderAllow: []string{"ops"}})
	assert.True(t, flag(t, out, "found"))
	assert.Equal(t, "@host:db1", text(t, out, "id"))
	assert.True(t, flag(t, out, "host"))
	assert.False(t, flag(t, out, "shell"))
	assert.Equal(t, 0, list(t, out, "allow").Len())
	assert.Equal(t, 1, list(t, out, "folder_allow").Len())

	none := job(Job{})
	assert.False(t, flag(t, none, "found"))
	assert.Equal(t, "", text(t, none, "id"))
}

// An empty list is an empty array, so a foreach over it draws the empty line.
func TestFolders_Empty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, folders(nil).Len())
	assert.Equal(t, 0, running(nil).Len())
	assert.Equal(t, 0, texts(nil).Len())
}

func TestRunning(t *testing.T) {
	t.Parallel()

	out := running([]Running{{Key: "infra/build", ID: "01hz", State: "running"}})
	assert.Equal(t, 1, out.Len())

	row := at(t, out, 0)
	assert.Equal(t, "infra/build", text(t, row, "key"))
	assert.Equal(t, "01hz", text(t, row, "id"))
	assert.Equal(t, "running", text(t, row, "state"))
}

// The six keys are what Task\Permission reads, so a rename here is a page that
// decides nothing.
func TestRules(t *testing.T) {
	t.Parallel()

	out := rules(Rules{
		Rules:    []Rule{{Key: "view", Title: "Allow View"}},
		Sections: []Rule{{Key: "note", Title: "Notes"}},
		Paths:    []string{"note/all", "note/view"},
		Checks:   []string{"note/view"},
	})
	for _, name := range []string{"rules", "sections", "permissions", "paths", "checks", "extras"} {
		_, ok := out.Get(name)
		assert.True(t, ok, name)
	}

	assert.Equal(t, 2, list(t, out, "paths").Len())
	assert.Equal(t, 1, list(t, out, "checks").Len())
	// Even the lists nothing filled are arrays, so a count over them is 0.
	assert.Equal(t, 0, list(t, out, "permissions").Len())
	assert.Equal(t, 0, list(t, out, "extras").Len())

	assert.Equal(t, "note", text(t, at(t, list(t, out, "sections"), 0), "key"))
	assert.Equal(t, "Notes", text(t, at(t, list(t, out, "sections"), 0), "title"))
}

func TestPhpMap(t *testing.T) {
	t.Parallel()

	out := phpMap("a", 1, "b", 2)
	assert.Equal(t, 2, out.Len())
	assert.Equal(t, "x", text(t, phpMap("k", "x"), "k"))

	// A trailing key with no value is left out and not written null.
	assert.Equal(t, 1, phpMap("a", 1, "b").Len())
}

func TestPhpList(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 3, phpList("a", "b", "c").Len())
	assert.Equal(t, 0, phpList().Len())
}

func TestTitles(t *testing.T) {
	t.Parallel()

	out := titles([]Rule{{Key: "view", Title: "Allow View"}})
	assert.Equal(t, 1, out.Len())
	assert.Equal(t, "Allow View", text(t, at(t, out, 0), "title"))
}

func TestTexts(t *testing.T) {
	t.Parallel()

	out := texts([]string{"a", "b"})
	assert.Equal(t, 2, out.Len())

	var seen []string
	out.Range(func(_, value any) bool {
		seen = append(seen, value.(string))
		return true
	})
	assert.Equal(t, []string{"a", "b"}, seen, "in the order given")
}
