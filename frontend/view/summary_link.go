package view

import (
	"strings"
	"unicode/utf8"
)

// summaryPreview is how much of a summary the Files list shows on
// hover, in runes.
const summaryPreview = 240

// SummaryLink is the summary Claude wrote of a run, as a row of the
// Files list links it: the run page's summary, and the start of its
// text for the hover.
type SummaryLink struct {
	Href    string
	Preview string
}

// NewSummaryLink links the summary thread of a run, its reply's text
// cut to the preview on one line.
func NewSummaryLink(runID, threadID, text string) *SummaryLink {
	preview := strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(preview) > summaryPreview {
		preview = string([]rune(preview)[:summaryPreview]) + "…"
	}
	return &SummaryLink{Href: "/runs/" + runID + "#note-" + threadID, Preview: preview}
}
