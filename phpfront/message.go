package phpfront

import (
	"net/url"
	"regexp"
	"strconv"
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

// names reads a PHP array of strings, whatever the script made it with.
//
// The parameter is any and not *model.Array because the runtime has two
// spellings of a list: an array a script wrote is a *model.Array, and what
// array_keys, array_values or explode answer is a Go slice. A binding that
// declares *model.Array refuses the slice with a TypeError that reads
// "Argument #1 must be of type array, array given", and a list built by one
// of those functions is the ordinary way a page asks about many rows at
// once - Task\Folders::warm asks with array_keys, and the tree it draws is
// on every page.
//
// model.RangeValues is the runtime's own reader of both, so this takes
// whatever it takes and a value that is no list reads as no names.
func names(list any) []string {
	size, _ := model.LenValues(list)
	out := make([]string, 0, size)
	model.RangeValues(list, func(_, value any) bool {
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			out = append(out, name)
		}
		return true
	})
	return out
}

// fields reads a PHP array as the map a Go caller takes, keyed the way PHP
// keys it: a string key as itself and a list index as its decimal. It takes
// any for the same reason names does.
func fields(array any) map[string]any {
	size, _ := model.LenValues(array)
	out := make(map[string]any, size)
	model.RangeValues(array, func(key, value any) bool {
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
