package access_test

import (
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	hash, err := access.HashPassword("secret")
	assert.NoError(t, err)
	assert.True(t, access.Hashed(hash))
	assert.True(t, access.CheckHash(hash, "secret"))
	assert.False(t, access.CheckHash(hash, "wrong"))
}

func TestCheckHash(t *testing.T) {
	t.Parallel()
	assert.False(t, access.CheckHash("", "anything"))
	assert.False(t, access.CheckHash("not a hash", "anything"))
}

func TestHashed(t *testing.T) {
	t.Parallel()
	assert.True(t, access.Hashed("$2a$10$abc"))
	assert.False(t, access.Hashed("secret"))
	assert.False(t, access.Hashed("$2"))
}
