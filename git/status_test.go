package git_test

import (
	"testing"

	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
)

// ParseStatus reads the branch headers and counts the entries.
func TestParseStatus(t *testing.T) {
	t.Parallel()
	s := git.ParseStatus("# branch.oid 0123456789abcdef\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +3 -1\n1 .M N... 100644 100644 100644 a a a\n? new\n")
	assert.Equal(t, &git.Status{Branch: "main", Upstream: "origin/main", Ahead: 3, Behind: 1, Dirty: 2}, s)
	s = git.ParseStatus("# branch.oid 0123456789abcdef\n# branch.head (detached)\n")
	assert.Equal(t, "0123456", s.Detached)
	assert.Equal(t, &git.Status{Branch: "main"}, git.ParseStatus("# branch.oid (initial)\n# branch.head main\n"))
}
