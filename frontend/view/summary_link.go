package view

import (
	"strings"
	"unicode/utf8"

	"github.com/titpetric/task-ui/note"
)

// summaryPreview is how much of a summary the Files list shows on
// hover, in runes.
const summaryPreview = 240

// SummaryLink is the summary Claude wrote of a run, as a row of the
// Files list links it: the run's report, and the start of its text for
// the hover.
type SummaryLink struct {
	Href    string
	Preview string
}

// NewSummaryLink links the report of a run, its summary's text cut to
// the preview on one line.
func NewSummaryLink(runID, text string) *SummaryLink {
	preview := strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(preview) > summaryPreview {
		preview = string([]rune(preview)[:summaryPreview]) + "…"
	}
	return &SummaryLink{Href: (&note.Report{RunID: runID}).Href(), Preview: preview}
}
