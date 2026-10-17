package access_test

import (
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	hash, err := access.HashPassword("secret")
	assert.NoError(t, err)
	assert.True(t, access.Hashed(hash))
	user := &model.User{Name: "a", Password: hash}
	assert.True(t, access.CheckPassword(user, "secret"))
	assert.False(t, access.CheckPassword(user, "wrong"))
}

func TestCheckPassword(t *testing.T) {
	t.Parallel()
	assert.False(t, access.CheckPassword(nil, "anything"))
	assert.False(t, access.CheckPassword(&model.User{Password: "not a hash"}, "anything"))
}

func TestHashed(t *testing.T) {
	t.Parallel()
	assert.True(t, access.Hashed("$2a$10$abc"))
	assert.False(t, access.Hashed("secret"))
	assert.False(t, access.Hashed("$2"))
}
