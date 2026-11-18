package phpfront

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
)

// The uptime reads as a person counts it: the units that passed and none of
// the ones that did not, so an instance up for an hour shows the hour and not
// a day of zeros in front of it.
func TestUptime(t *testing.T) {
	t.Parallel()

	cases := []struct {
		run    time.Duration
		worded string
	}{
		{0, "0s"},
		{500 * time.Millisecond, "0s"},
		{-time.Minute, "0s"},
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m 30s"},
		{time.Hour, "1h"},
		{4*time.Hour + 7*time.Minute, "4h 7m"},
		{25 * time.Hour, "1d 1h"},
		{26*time.Hour + 3*time.Minute + 4*time.Second, "1d 2h 3m 4s"},
	}

	for _, c := range cases {
		assert.Equal(t, c.worded, Uptime(c.run))
	}
}
