package retention_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// items are one a day for ten days, oldest first, each 10 bytes.
func items(now time.Time) []retention.Item {
	var out []retention.Item
	for day := 10; day >= 1; day-- {
		out = append(out, retention.Item{ID: string(rune('a' + 10 - day)), At: now.Add(-time.Duration(day)*24*time.Hour + time.Minute), Size: 10})
	}
	return out
}

// Each age lets the older items go; the limit takes the oldest until
// the rest fit; both together take every item either takes; never with
// no limit keeps everything.
func TestExpire(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	all := items(now)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Day}, now), 9)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Week}, now), 3)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Month}, now), 0)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Never}, now), 0)

	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Never, Limit: 40}, now), 6)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Never, Limit: 45}, now), 6)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Never, Limit: 100}, now), 0)

	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Week, Limit: 40}, now), 6)
	assert.Len(t, retention.Expire(all, retention.Rule{Age: retention.Week, Limit: 90}, now), 3)
	assert.Len(t, retention.Expire(nil, retention.Rule{Age: retention.Day, Limit: 1}, now), 0)

	// A busy item stays and counts towards the limit.
	busy := items(now)
	busy[0].Busy = true
	gone := retention.Expire(busy, retention.Rule{Age: retention.Never, Limit: 80}, now)
	assert.Len(t, gone, 2)
	assert.Equal(t, "b", gone[0].ID)
}
