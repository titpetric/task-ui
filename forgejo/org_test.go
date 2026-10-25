package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A workspace takes the organisation's full name, else its name.
func TestOrg(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Malta d.o.o.", forgejo.Org{Name: "malta", FullName: "Malta d.o.o."}.Title())
	assert.Equal(t, "malta", forgejo.Org{Name: "malta"}.Title())
}
