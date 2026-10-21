package folder_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// Sealed takes the repository, the credential files and any path
// through them, and none of their look-alikes.
func TestSealed(t *testing.T) {
	t.Parallel()
	for _, name := range []string{".git", ".env", ".env.local", ".env.production", ".htpasswd", ".htaccess"} {
		assert.True(t, folder.Sealed(name), name)
	}
	for _, name := range []string{".gitignore", ".github", ".envrc", "env", ".claude", "README.md", ".atkins"} {
		assert.False(t, folder.Sealed(name), name)
	}
	assert.True(t, folder.SealedPath(".git/config"))
	assert.True(t, folder.SealedPath("sub/.env.local"))
	assert.True(t, folder.SealedPath("a/.git/b/c"))
	assert.False(t, folder.SealedPath("a/.gitignore"))
	assert.False(t, folder.SealedPath(".claude/settings.json"))
}
