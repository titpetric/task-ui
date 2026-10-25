package retention_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// Each age keeps for its time, never for ever; a rule says itself, and
// the page's labels name every age and every listed size.
func TestRule(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 24*time.Hour, retention.Rule{Age: retention.Day}.Keep())
	assert.Equal(t, 7*24*time.Hour, retention.Rule{Age: retention.Week}.Keep())
	assert.Equal(t, 31*24*time.Hour, retention.Rule{Age: retention.Month}.Keep())
	assert.Equal(t, 90*24*time.Hour, retention.Rule{Age: retention.Quarter}.Keep())
	assert.Equal(t, 365*24*time.Hour, retention.Rule{Age: retention.Year}.Keep())
	assert.Equal(t, time.Duration(0), retention.Rule{Age: retention.Never}.Keep())
	assert.True(t, retention.Rule{Age: retention.Never}.Valid())
	assert.True(t, retention.Rule{Age: retention.Year}.Valid())
	assert.False(t, retention.Rule{Age: "3d"}.Valid())
	assert.False(t, retention.Rule{Age: retention.Day, Limit: -1}.Valid())
	assert.Equal(t, "keep for 31 days, at most 1.0 GB", retention.Rule{Age: retention.Month, Limit: 1 << 30}.Says())
	assert.Equal(t, "keep for 90 days, no size limit", retention.Rule{Age: retention.Quarter}.Says())
	assert.Equal(t, "keep for 365 days", retention.Rule{Age: retention.Year}.Keeps())
	assert.Equal(t, "keep for ever, no size limit", retention.Rule{Age: retention.Never}.Says())
	assert.Equal(t, []string{"24h", "7d", "31d", "90d", "365d", "never"}, retention.Ages)
	assert.Equal(t, "90 days", retention.Label(retention.Quarter))
	assert.Equal(t, "Never", retention.Label(retention.Never))
	assert.Equal(t, "500 MB", retention.SizeLabel(500<<20))
	assert.Equal(t, "50 GB", retention.SizeLabel(50<<30))
	assert.Equal(t, "1.5 GB", retention.SizeLabel(3<<29))
	assert.Len(t, retention.Limits, 6)
}
