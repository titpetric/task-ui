package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A user is an owner of the user kind, named by the login and titled by
// the full name.
func TestUser(t *testing.T) {
	t.Parallel()
	owner := forgejo.User{Login: "black", FullName: "Tit Petric"}.Owner()
	assert.Equal(t, forgejo.Owner{Name: "black", FullName: "Tit Petric", Kind: forgejo.OwnerUser}, owner)
	assert.Equal(t, "Tit Petric", owner.Title())
	assert.Equal(t, "black", forgejo.User{Login: "black"}.Owner().Title())
}
