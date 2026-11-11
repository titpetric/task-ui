package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// An organisation is an owner of the org kind, with its names and
// description.
func TestOrg(t *testing.T) {
	t.Parallel()
	owner := forgejo.Org{Name: "malta", FullName: "Malta d.o.o.", Description: "The firm"}.Owner()
	assert.Equal(t, forgejo.Owner{Name: "malta", FullName: "Malta d.o.o.", Description: "The firm", Kind: forgejo.OwnerOrg}, owner)
	assert.Equal(t, "Malta d.o.o.", owner.Title())
}
