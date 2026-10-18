package access_test

import (
	"regexp"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestBackupPasswords(t *testing.T) {
	t.Parallel()
	passwords, hashes, err := access.BackupPasswords(3)
	assert.NoError(t, err)
	assert.Len(t, passwords, 3)
	assert.Len(t, hashes, 3)
	shape := regexp.MustCompile(`^[a-z2-9]{4}-[a-z2-9]{4}-[a-z2-9]{4}$`)
	seen := map[string]bool{}
	for i, p := range passwords {
		assert.True(t, shape.MatchString(p), p)
		assert.True(t, access.CheckHash(hashes[i], p))
		assert.False(t, seen[p])
		seen[p] = true
	}

	for _, bad := range []int{0, -1, 101} {
		_, _, err := access.BackupPasswords(bad)
		assert.Error(t, err)
	}
}
