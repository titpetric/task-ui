package view

import (
	"html/template"
	"net/url"
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
	return message(text, func(string) bool { return false })
}

// Message renders the text of a note as the Message function does, and
// links a mention of a user to their profile when the signed in user may
// open it.
func (p *Page) Message(text string) template.HTML {
	return message(text, p.CanProfile)
}

// message renders a note's text, linking the mentions that link names.
func message(text string, link func(name string) bool) template.HTML {
	return template.HTML(markdown.Render(markdown.Options{
		HardBreaks: true,
		Text: func(escaped string) string {
			return mentionText.ReplaceAllStringFunc(escaped, func(m string) string {
				sub := mentionText.FindStringSubmatch(m)
				lead, name := sub[1], sub[2]
				if link(name) {
					return lead + `<a class="mention" href="/users/` + url.PathEscape(name) + `">@` + name + `</a>`
				}
				return lead + `<span class="mention">@` + name + `</span>`
			})
		},
	}, []byte(text)))
}
