package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// A block estimates its limit by the average item, and lists the ages
// with the rule's checked.
func TestRetentionBlock(t *testing.T) {
	t.Parallel()
	b := view.RetentionBlock{One: "run", Many: "runs", Usage: retention.Usage{Count: 10, Bytes: 10 << 20}, Rule: retention.Rule{Age: retention.Week, Limit: 100 << 20}}
	assert.Equal(t, int64(1<<20), b.Average())
	assert.Equal(t, "about 100 runs", b.Estimate())
	assert.Equal(t, "100.0 MB", b.LimitText())
	ages := b.Ages()
	assert.Len(t, ages, 4)
	assert.Equal(t, "7 days", ages[1].Label)
	assert.True(t, ages[1].Checked)
	assert.False(t, ages[0].Checked)
	assert.Len(t, b.Sizes(), 3)

	empty := view.RetentionBlock{Rule: retention.Rule{Age: retention.Never}}
	assert.Equal(t, int64(0), empty.Average())
	assert.Equal(t, "", empty.Estimate())
	assert.Equal(t, "", empty.LimitText())
}
