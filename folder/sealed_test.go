package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// Sealed takes .git and every dot file, and leaves the dot directories
// other than .git and what is in them open.
func TestSealed(t *testing.T) {
	t.Parallel()
	for _, name := range []string{".env", ".env.local", ".htpasswd", ".htaccess", ".gitignore", ".envrc", ".git", ".x"} {
		assert.True(t, folder.Sealed(name, false), name)
	}
	assert.True(t, folder.Sealed(".git", true))
	for _, name := range []string{".github", ".claude", ".atkins", "env", "bin"} {
		assert.False(t, folder.Sealed(name, true), name)
	}
	for _, name := range []string{"README.md", "env", "Makefile"} {
		assert.False(t, folder.Sealed(name, false), name)
	}
	assert.True(t, folder.SealedPath(".git/config", false))
	assert.True(t, folder.SealedPath("a/.git/b/c", true))
	assert.True(t, folder.SealedPath("sub/.env.local", false))
	assert.True(t, folder.SealedPath(".claude/.secret", false))
	assert.True(t, folder.SealedPath("./.foo", false))
	assert.False(t, folder.SealedPath(".claude/CLAUDE.md", false))
	assert.False(t, folder.SealedPath(".atkins/skills/x.yml", false))
	assert.False(t, folder.SealedPath(".claude", true))
	assert.False(t, folder.SealedPath("", true))
}
