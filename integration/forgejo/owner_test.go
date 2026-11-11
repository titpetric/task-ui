package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A workspace takes the owner's full name, else its name.
func TestOwner(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Malta d.o.o.", forgejo.Owner{Name: "malta", FullName: "Malta d.o.o."}.Title())
	assert.Equal(t, "malta", forgejo.Owner{Name: "malta"}.Title())
}
