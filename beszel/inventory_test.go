package beszel_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// Hub is the client of the settings, or ErrOff without them.
func TestInventory_Hub(t *testing.T) {
	t.Parallel()
	inv := &beszel.Inventory{Settings: func(context.Context) (string, string, error) { return "http://hub", "t", nil }}
	a, err := inv.Hub(context.Background())
	assert.NoError(t, err)
	b, _ := inv.Hub(context.Background())
	assert.True(t, a == b)
}
