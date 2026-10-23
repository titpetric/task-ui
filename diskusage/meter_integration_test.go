//go:build integration

package diskusage_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/internal/assert"
)

// The meter keeps a measure for the TTL, and Forget drops it.
func TestMeter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, "a.txt", 10)
	m := &diskusage.Meter{TTL: time.Hour}
	assert.Equal(t, int64(10), m.Measure(dir).Files)

	write(t, dir, "b.txt", 5)
	assert.Equal(t, int64(10), m.Measure(dir).Files, "kept")
	m.Forget(dir)
	assert.Equal(t, int64(15), m.Measure(dir).Files, "measured again")
}
