package retention_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// Each age keeps for its time, never for ever; a rule says itself.
func TestRule(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 24*time.Hour, retention.Rule{Age: retention.Day}.Keep())
	assert.Equal(t, 7*24*time.Hour, retention.Rule{Age: retention.Week}.Keep())
	assert.Equal(t, 31*24*time.Hour, retention.Rule{Age: retention.Month}.Keep())
	assert.Equal(t, time.Duration(0), retention.Rule{Age: retention.Never}.Keep())
	assert.True(t, retention.Rule{Age: retention.Never}.Valid())
	assert.False(t, retention.Rule{Age: "3d"}.Valid())
	assert.False(t, retention.Rule{Age: retention.Day, Limit: -1}.Valid())
	assert.Equal(t, "keep for 31 days, at most 1.0 GB", retention.Rule{Age: retention.Month, Limit: 1 << 30}.Says())
	assert.Equal(t, "keep for ever, no size limit", retention.Rule{Age: retention.Never}.Says())
	assert.Len(t, retention.Ages, 4)
}
