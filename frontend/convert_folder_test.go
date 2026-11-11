package phpfront

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The keys a folder's page indexes by. The helpers are convert_test.go's.

// A place is a folder or a directory inside one, and both answer the same
// keys: the page draws one template over either, so the directory's answer
// carries the folder's gate in allow and its own path in rel.
func TestPlace(t *testing.T) {
	t.Parallel()

	out := place(Place{
		Found: true,
		ID:    "infra",
		Title: "Infra",
		Dir:   "/src/infra",
		Rel:   "api",
		Allow: []string{"ops"},
		Hours: "weekdays 09-17",
		Sources: []Source{{
			Kind: "task", File: "Taskfile.yml", Command: []string{"task"}, Lines: 42,
		}},
		Includes: []File{{Name: "Taskfile.dist.yml", Lines: 12}},
		Errors:   []Finding{{Path: "Taskfile.yml", Job: "build", Issue: "no command"}},
	})

	assert.True(t, flag(t, out, "found"))
	assert.Equal(t, "infra", text(t, out, "id"))
	assert.Equal(t, "api", text(t, out, "rel"), "the directory inside, which the breadcrumbs walk")
	assert.Equal(t, "weekdays 09-17", text(t, out, "hours"))
	assert.Equal(t, 1, list(t, out, "allow").Len())

	source := at(t, list(t, out, "sources"), 0)
	assert.Equal(t, "task", text(t, source, "kind"))
	assert.Equal(t, "Taskfile.yml", text(t, source, "file"))
	assert.Equal(t, 1, list(t, source, "command").Len())
	assert.Equal(t, 42, value[int](t, source, "lines"), "what the button that shows the file says")

	include := at(t, list(t, out, "includes"), 0)
	assert.Equal(t, "Taskfile.dist.yml", text(t, include, "name"))
	assert.Equal(t, 12, value[int](t, include, "lines"))

	finding := at(t, list(t, out, "errors"), 0)
	assert.Equal(t, "build", text(t, finding, "job"))
	assert.Equal(t, "no command", text(t, finding, "issue"))

	// A place nothing found answers every key, so the page gates on found
	// and not on a key being absent.
	none := place(Place{})
	assert.False(t, flag(t, none, "found"))
	assert.Equal(t, 0, list(t, none, "sources").Len())
	assert.Equal(t, 0, list(t, none, "errors").Len())
}

// A row of the Jobs table carries the folder its jobs run as, which is the
// whole of their gate: the row does not state a rule of its own, so a template
// reads folder.allow and the job's allow and nothing else.
func TestGroups(t *testing.T) {
	t.Parallel()

	out := groups([]Group{{
		Kind:    "task",
		Note:    "Taskfile.yml, task",
		Source:  "Taskfile.yml",
		Dir:     "api",
		Targets: true,
		Folder:  Where{ID: "infra", Rel: "api", Allow: []string{"ops"}, Hours: "weekdays 09-17"},
		Jobs: []Job{{
			ID: "build", Name: "build", Kind: "task", Dir: "api",
			Description: "build the image", Key: "infra/build",
			Interactive: true, TakesArgs: true, ArgsCommand: "task build --",
			Allow: []string{"ann"}, Confirm: "this rebuilds from scratch",
		}},
	}, {
		Kind:  "task",
		Empty: "Taskfile.yml declares no targets",
	}})
	assert.Equal(t, 2, out.Len())

	row := at(t, out, 0)
	assert.Equal(t, "task", text(t, row, "kind"))
	assert.Equal(t, "Taskfile.yml, task", text(t, row, "note"))
	assert.Equal(t, "api", text(t, row, "dir"), "the row's label links to the directory inside")
	assert.True(t, flag(t, row, "targets"))

	folder := list(t, row, "folder")
	assert.Equal(t, "infra", text(t, folder, "id"))
	assert.Equal(t, 1, list(t, folder, "allow").Len())
	assert.Equal(t, "weekdays 09-17", text(t, folder, "hours"))

	job := at(t, list(t, row, "jobs"), 0)
	assert.Equal(t, "infra/build", text(t, job, "key"), "what a pin and a schedule are written with")
	assert.Equal(t, "build the image", text(t, job, "description"))
	assert.True(t, flag(t, job, "interactive"))
	assert.True(t, flag(t, job, "takes_args"))
	assert.Equal(t, "task build --", text(t, job, "args_command"))
	assert.Equal(t, "this rebuilds from scratch", text(t, job, "confirm"))
	assert.Equal(t, 1, list(t, job, "allow").Len())

	// A source with no targets is a row that says so and lists nothing, not
	// a row left out: the page states what it read.
	assert.Equal(t, "Taskfile.yml declares no targets", text(t, at(t, out, 1), "empty"))
	assert.Equal(t, 0, list(t, at(t, out, 1), "jobs").Len())
}

// A directory of the Files panel answers whether it may be opened at all: a
// path is empty for an entry the folder refuses, which is what the row reads
// instead of composing the link itself.
func TestEntries(t *testing.T) {
	t.Parallel()

	out := entries([]Entry{
		{Name: "nested", Dir: true, Size: 4096, ModTime: "2026-11-02 10:32:12.414 +0000 UTC", Path: "nested", Folder: "infra"},
		{Name: "run.sh", Size: 31, Path: "run.sh", Folder: "infra", Exec: true},
		{Name: ".env", Size: 64, Sealed: true},
	})
	assert.Equal(t, 3, out.Len())

	dir := at(t, out, 0)
	assert.True(t, flag(t, dir, "dir"))
	assert.Equal(t, int64(4096), value[int64](t, dir, "size"))
	assert.Equal(t, "2026-11-02 10:32:12.414 +0000 UTC", text(t, dir, "mod_time"))
	assert.Equal(t, "infra", text(t, dir, "folder"))

	assert.True(t, flag(t, at(t, out, 1), "exec"))

	sealed := at(t, out, 2)
	assert.True(t, flag(t, sealed, "sealed"))
	assert.Equal(t, "", text(t, sealed, "path"), "no path is an entry the folder will not open")

	assert.Equal(t, 0, entries(nil).Len())
}

// What git reports of a checkout, with the two-letter code of every change
// beside what it says in words: the page prints the words and sorts on the
// code, so both cross.
func TestStatus(t *testing.T) {
	t.Parallel()

	out := status(Status{
		Found:    true,
		Checkout: "/src/infra",
		Branch:   "main",
		Upstream: "origin/main",
		Ahead:    2,
		Behind:   1,
		Dirty:    3,
		Changes: []Change{
			{Code: " M", State: "modified", Path: "api/main.go"},
			{Code: "R ", State: "renamed", Path: "api/run.go", From: "api/start.go"},
		},
	})

	assert.True(t, flag(t, out, "found"))
	assert.Equal(t, "main", text(t, out, "branch"))
	assert.Equal(t, "origin/main", text(t, out, "upstream"))
	assert.Equal(t, 2, value[int](t, out, "ahead"))
	assert.Equal(t, 1, value[int](t, out, "behind"))
	assert.Equal(t, 3, value[int](t, out, "dirty"))

	assert.Equal(t, " M", text(t, at(t, list(t, out, "changes"), 0), "code"))
	assert.Equal(t, "modified", text(t, at(t, list(t, out, "changes"), 0), "state"))
	assert.Equal(t, "api/start.go", text(t, at(t, list(t, out, "changes"), 1), "from"), "where a rename came from")

	// A folder in no checkout is found false with an empty changes array,
	// not a missing key, so the panel draws its empty line.
	assert.False(t, flag(t, status(Status{}), "found"))
	assert.Equal(t, 0, list(t, status(Status{}), "changes").Len())
}

// The commit a folder is at, with the space it takes nested inside: the page
// draws both in one line, and nothing having measured the space is a measure
// of false and not of zero.
func TestCheckout(t *testing.T) {
	t.Parallel()

	out := checkout(Checkout{
		Found:   true,
		Remote:  "git@github.com:malta/infra",
		Branch:  "main",
		Commit:  "9bbee11",
		Subject: "feat: the first page",
		When:    "2026-11-02 10:32:12.414 +0000 UTC",
		Space:   Space{Measured: true, Files: 2048, History: 512},
	})

	assert.Equal(t, "9bbee11", text(t, out, "commit"))
	assert.Equal(t, "feat: the first page", text(t, out, "subject"))
	assert.True(t, flag(t, list(t, out, "space"), "measured"))
	assert.Equal(t, int64(2048), value[int64](t, list(t, out, "space"), "files"))

	assert.False(t, flag(t, checkout(Checkout{}), "found"))
	assert.False(t, flag(t, list(t, checkout(Checkout{}), "space"), "measured"))
}

// An hours rule is the one line a page prints and whether it is open now. A
// rule the parser refuses prints as itself and is open, as a run of it is
// never refused.
func TestHours(t *testing.T) {
	t.Parallel()

	out := hours(Hours{Line: "weekdays 09-17", Open: true})
	assert.Equal(t, "weekdays 09-17", text(t, out, "line"))
	assert.True(t, flag(t, out, "open"))

	assert.False(t, flag(t, hours(Hours{Line: "weekdays 09-17"}), "open"))
}
