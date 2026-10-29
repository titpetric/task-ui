package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// Started is a time a script can count from: milliseconds since the
// epoch, and an empty string where there is none.
//
// It exists because Duration's text cannot be read back. The runs page
// ticked the duration of a running row by parsing the rendered value, so
// a row at "2m05s" parsed as 2 and a row at "1h02m" as 1, and the column
// restarted near zero on the first tick after the page loaded.
func TestStarted(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.October, 29, 22, 0, 0, 0, time.UTC)

	assert.Equal(t, "1793311200000", view.Started(at))
	assert.Equal(t, "1793311200000", view.Started(&at))
	assert.Equal(t, "", view.Started(nil))
	assert.Equal(t, "", view.Started(time.Time{}))
	assert.Equal(t, "", view.Started((*time.Time)(nil)))
	assert.Equal(t, "", view.Started("not a time"))

	// The value is what the page writes, and what the ticker reads back.
	assert.Equal(t, "2m05s", view.Duration(125*time.Second))
	assert.Equal(t, "1h02m", view.Duration(62*time.Minute))
}
