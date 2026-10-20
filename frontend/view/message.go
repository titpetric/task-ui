package view

import (
	"html/template"
	"regexp"

	"github.com/titpetric/task-ui/markdown"
)

// mentionText matches @name where a note names a user, the way
// model.Mentions reads it.
var mentionText = regexp.MustCompile(`(^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*[A-Za-z0-9_]|[A-Za-z0-9])`)

// Message renders the text of a note as markdown: lists, **bold**,
// headings, `code` and ``` fences, with a line break kept a line break
// and @name marked as a mention. Raw HTML in the text shows as text.
func Message(text string) template.HTML {
	return template.HTML(markdown.Render(markdown.Options{
		HardBreaks: true,
		Text: func(escaped string) string {
			return mentionText.ReplaceAllString(escaped, `$1<span class="mention">@$2</span>`)
		},
	}, []byte(text)))
}
