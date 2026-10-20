package view

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

// inlineCode matches `code` on one line.
var inlineCode = regexp.MustCompile("`([^`\n]+)`")

// mentionText matches @name where a note names a user, the way
// model.Mentions reads it.
var mentionText = regexp.MustCompile(`(^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*[A-Za-z0-9_]|[A-Za-z0-9])`)

// Message renders the text of a note: a ``` fence is a code block, `code`
// inline code, @name a mention, and a line break stays a line break.
// Everything else is text, escaped.
func Message(text string) template.HTML {
	var out strings.Builder
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "```")
	if len(parts)%2 == 0 {
		// A fence that never closes is text.
		last := len(parts) - 1
		parts[last-1] += "```" + parts[last]
		parts = parts[:last]
	}
	for i, part := range parts {
		if i%2 == 1 {
			// A fence's first line may name the language; the block
			// keeps the rest.
			if nl := strings.IndexByte(part, '\n'); nl >= 0 && !strings.ContainsAny(part[:nl], " \t") {
				part = part[nl+1:]
			}
			out.WriteString(`<pre class="note-code">` + html.EscapeString(strings.Trim(part, "\n")) + "</pre>")
			continue
		}
		out.WriteString(prose(strings.Trim(part, "\n")))
	}
	return template.HTML(out.String())
}

// prose renders text outside code blocks.
func prose(text string) string {
	if text == "" {
		return ""
	}
	escaped := html.EscapeString(text)
	escaped = inlineCode.ReplaceAllString(escaped, "<code>$1</code>")
	escaped = mentionText.ReplaceAllString(escaped, `$1<span class="mention">@$2</span>`)
	paragraphs := strings.Split(escaped, "\n\n")
	for i, p := range paragraphs {
		paragraphs[i] = "<p>" + strings.ReplaceAll(p, "\n", "<br>") + "</p>"
	}
	return strings.Join(paragraphs, "")
}
