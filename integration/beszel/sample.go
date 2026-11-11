package beszel

import (
	"time"
)

// Sample is a record of a host's measures over a span of the hub's
// longer records: when the hub wrote it, and the averages of the span.
type Sample struct {
	At    time.Time
	Stats Stats
}
