package view

import (
	"html/template"

	"github.com/titpetric/task-ui/markdown"
)

// Markdown renders Claude's markdown, a summary or a reply, as HTML.
func Markdown(text string) template.HTML {
	// Raw HTML in it shows as text, and links keep to safe schemes.
	out := markdown.HTML([]byte(text), "", "")
	return template.HTML(out)
}
