package phpfront

import (
	"fmt"
	"strings"
	"time"
)

// Stamp is how the Instance page prints when the server started: the local
// time of the server's clock, to the second, with the zone's name.
const Stamp = "2006-01-02 15:04:05 MST"

// Uptime words how long the server has been up as the Instance page prints
// it: the units from days down, largest first, and no unit that is zero, so
// an instance up for four hours reads "4h 7m" and not "0d 4h 7m 0s". A
// duration under a second, and anything before the start, reads "0s".
func Uptime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	seconds := int64(d / time.Second)
	days := seconds / 86400
	hours := seconds % 86400 / 3600
	minutes := seconds % 3600 / 60
	secs := seconds % 60
	parts := make([]string, 0, 4)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if secs > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}
	return strings.Join(parts, " ")
}
