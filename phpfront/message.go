package phpfront

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/titpetric/phpscript/modelutil"

	"github.com/titpetric/task-ui/internal/markdown"
)

// mention matches @name where a note names a user; a colon right after the name makes it a job ID such as @host:dev1 and no mention.
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

// names reads a PHP array of strings, whatever spelling of a list it holds.
func names(list any) []string {
	size, _ := modelutil.LenValues(list)
	out := make([]string, 0, size)
	modelutil.RangeValues(list, func(_, value any) bool {
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			out = append(out, name)
		}
		return true
	})
	return out
}

// fields reads a PHP array as a map, a list index keyed as its decimal.
func fields(array any) map[string]any {
	size, _ := modelutil.LenValues(array)
	out := make(map[string]any, size)
	modelutil.RangeValues(array, func(key, value any) bool {
		switch k := key.(type) {
		case string:
			out[k] = value
		case int:
			out[strconv.Itoa(k)] = value
		case int64:
			out[strconv.FormatInt(k, 10)] = value
		}
		return true
	})
	return out
}

// pairs reads a PHP array as a map of strings, dropping a value of any other type.
func pairs(array any) map[string]string {
	out := make(map[string]string)
	for key, value := range fields(array) {
		if text, ok := value.(string); ok {
			out[key] = text
		}
	}
	return out
}
