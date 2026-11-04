//go:build integration

package app

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// An instance with no artifact tree and no meter says nothing about space,
// rather than that the tree takes none.
//
// Here and not in the unit suite because Tree reaches diskusage.Measure, which
// runs a process: this asserts the guard in front of it, and the guard is only
// worth anything if the call behind it is the real one.
func TestFrontHost_Space(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Tree(context.Background(), ""))
	assert.False(t, (&FrontHost{}).Space(context.Background()).Measured)
}
