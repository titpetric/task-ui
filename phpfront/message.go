package phpfront

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/titpetric/phpscript/model"

	"github.com/titpetric/task-ui/markdown"
)

// mention matches @name where a note names a user, the way model.Mentions
// reads it. The third group, a colon and a character other than a space,
// makes it a job ID such as @host:dev1, which is no mention.
var mention = regexp.MustCompile(`(^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*[A-Za-z0-9_]|[A-Za-z0-9])((?::\S)?)`)

// Message renders the text of a note: markdown with a line break kept as one
// and @name marked, through the renderer that owns the content security
// policy, so raw HTML in the text shows as text.
//
// A name in link becomes a link under prefix, which carries the mount path.
// Every other mention is marked and not linked.
func Message(text, prefix string, link []string) string {
	linkable := make(map[string]bool, len(link))
	for _, name := range link {
		linkable[name] = true
	}
	return markdown.Render(markdown.Options{
		HardBreaks: true,
		Text: func(escaped string) string {
			return mention.ReplaceAllStringFunc(escaped, func(m string) string {
				sub := mention.FindStringSubmatch(m)
				lead, name, job := sub[1], sub[2], sub[3]
				if job != "" {
					return m
				}
				if linkable[name] {
					return lead + `<a class="mention" href="` + prefix + url.PathEscape(name) + `">@` + name + `</a>`
				}
				return lead + `<span class="mention">@` + name + `</span>`
			})
		},
	}, []byte(text))
}

// names reads a PHP array of strings.
func names(list *model.Array) []string {
	if list == nil {
		return nil
	}
	out := make([]string, 0, list.Len())
	list.Range(func(_, value any) bool {
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			out = append(out, name)
		}
		return true
	})
	return out
}
