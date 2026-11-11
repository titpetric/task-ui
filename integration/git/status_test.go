package git_test

import (
	"testing"

	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
)

// ParseStatus reads the branch headers and the changed files.
func TestParseStatus(t *testing.T) {
	t.Parallel()
	s := git.ParseStatus("# branch.oid 0123456789abcdef\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +3 -1\n1 .M N... 100644 100644 100644 a a a\n? new\n")
	assert.Equal(t, &git.Status{Branch: "main", Upstream: "origin/main", Ahead: 3, Behind: 1, Dirty: 2, Changes: []git.Change{
		{Index: ".", Worktree: "M", Path: "a"},
		{Index: "?", Worktree: "?", Path: "new"},
	}}, s)
	s = git.ParseStatus("# branch.oid 0123456789abcdef\n# branch.head (detached)\n")
	assert.Equal(t, "0123456", s.Detached)
	assert.Equal(t, &git.Status{Branch: "main", Changes: []git.Change{}}, git.ParseStatus("# branch.oid (initial)\n# branch.head main\n"))
}

// A renamed file names where it came from, an unmerged one is a
// conflict, and each reads as the table shows it.
func TestParseStatus_Changes(t *testing.T) {
	t.Parallel()
	s := git.ParseStatus("# branch.oid 0123456789abcdef\n" +
		"2 R. N... 100644 100644 100644 a a R100 new\told\n" +
		"u UU N... 100644 100644 100644 100644 a b c both\n" +
		"1 D. N... 100644 100644 100644 a a gone\n" +
		"1 AM N... 100644 100644 100644 a a added\n")
	assert.Equal(t, 4, s.Dirty)
	assert.Equal(t, []git.Change{
		{Index: "R", Worktree: ".", Path: "new", From: "old"},
		{Index: "U", Worktree: "U", Path: "both", Conflict: true},
		{Index: "D", Worktree: ".", Path: "gone"},
		{Index: "A", Worktree: "M", Path: "added"},
	}, s.Changes)
	assert.Equal(t, "R.", s.Changes[0].Code())
	assert.Equal(t, "renamed", s.Changes[0].State())
	assert.Equal(t, "conflicted", s.Changes[1].State())
	assert.Equal(t, "deleted", s.Changes[2].State())
	assert.Equal(t, "AM", s.Changes[3].Code())
	assert.Equal(t, "added", s.Changes[3].State())
}

// Under counts the changed files in a directory of the checkout, all of
// them for the checkout itself, and none for a sibling sharing a prefix.
func TestStatus_Under(t *testing.T) {
	t.Parallel()
	s := git.ParseStatus("# branch.oid 0123456789abcdef\n" +
		"1 .M N... 100644 100644 100644 a a app/main.go\n" +
		"? app/new.txt\n" +
		"? apps/other.txt\n" +
		"? top.txt\n")
	assert.Equal(t, 4, s.Under(""))
	assert.Equal(t, 2, s.Under("app"))
	assert.Equal(t, 1, s.Under("apps"))
	assert.Equal(t, 0, s.Under("docs"))
}
