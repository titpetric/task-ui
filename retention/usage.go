package retention

import (
	"time"
)

// Usage is what one kind of data holds: how many items, their bytes,
// and when the oldest is from.
type Usage struct {
	Count  int
	Bytes  int64
	Oldest *time.Time
}

// usageOf adds up the items, oldest first.
func usageOf(items []Item) Usage {
	usage := Usage{Count: len(items)}
	for _, item := range items {
		usage.Bytes += item.Size
	}
	if len(items) > 0 {
		oldest := items[0].At
		usage.Oldest = &oldest
	}
	return usage
}

// Says reads the usage: "12 items, 3.4 MB".
func (u Usage) Says(one, many string) string {
	return u.Name(one, many) + ", " + FormatSize(u.Bytes)
}

// Name counts the items by name: "1 run", "12 runs".
func (u Usage) Name(one, many string) string {
	if u.Count == 1 {
		return itoa(u.Count) + " " + one
	}
	return itoa(u.Count) + " " + many
}
