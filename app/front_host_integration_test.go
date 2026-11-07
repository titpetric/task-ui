//go:build integration

package app

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// An instance with no artifact tree and no meter leaves Space empty, which is
// not the same answer as a tree of zero bytes.
//
// In the integration suite because Tree reaches diskusage.Measure, which runs
// a process: this asserts the guard in front of that call, and the guard means
// nothing unless the call behind it is the real one.
func TestFrontHost_Space(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Tree(context.Background(), ""))
	assert.False(t, (&FrontHost{}).Space(context.Background()).Measured)
}
