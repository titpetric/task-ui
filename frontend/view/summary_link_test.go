package view_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The link goes to the run's report; the preview is one
// line, cut short.
func TestNewSummaryLink(t *testing.T) {
	t.Parallel()
	link := view.NewSummaryLink("r1", "## Result\n\nThe build  passed.")
	assert.Equal(t, view.SummaryLink{Href: "/runs/r1/report", Preview: "## Result The build passed."}, *link)

	long := view.NewSummaryLink("r1", strings.Repeat("é", 300))
	assert.Equal(t, 241, utf8.RuneCountInString(long.Preview))
	assert.True(t, strings.HasSuffix(long.Preview, "…"))
}
