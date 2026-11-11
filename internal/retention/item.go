package retention

import (
	"time"
)

// Item is what a rule keeps or lets go: a recording, a run, a thread, a
// version of an artifact, a read of a secret. A busy item always stays,
// such as the recording of a run in progress, and counts
// towards the limit. Size is zero in an area whose items have none.
type Item struct {
	ID   string
	At   time.Time
	Size int64
	Busy bool
}

// Expire returns the items, given oldest first, older than the rule keeps or over its limit, skipping busy ones.
func Expire(items []Item, rule Rule, now time.Time) []Item {
	var total int64
	for _, item := range items {
		total += item.Size
	}
	keep := rule.Keep()
	var gone []Item
	for _, item := range items {
		old := keep > 0 && item.At.Before(now.Add(-keep))
		over := rule.Limit > 0 && total > rule.Limit
		if !old && !over {
			break
		}
		if item.Busy {
			continue
		}
		total -= item.Size
		gone = append(gone, item)
	}
	return gone
}
