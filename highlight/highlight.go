// Package highlight draws the files a folder's pipeline reads as
// coloured HTML: YAML pipelines and Taskfiles, Makefiles, and anything
// else as escaped text. It marks a line's parts with spans whose classes
// the stylesheet colours: k a key or target, c a comment, s a string, n a
// number or a boolean, v an interpolation or a variable, p punctuation.
package highlight

import (
	"html"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
)

// HTML returns the file's text, escaped, with its parts marked by the
// convention its name says.
func HTML(name, src string) template.HTML {
	base := strings.ToLower(filepath.Base(name))
	var line func(string) string
	switch {
	case strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml"):
		line = yamlLine
	case base == "makefile" || base == "gnumakefile" || strings.HasSuffix(base, ".mk"):
		line = makeLine
	default:
		line = html.EscapeString
	}
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		lines[i] = line(l)
	}
	return template.HTML(strings.Join(lines, "\n"))
}

func span(class, text string) string {
	return `<span class="` + class + `">` + html.EscapeString(text) + `</span>`
}

var (
	// yamlKey is a mapping key at the start of a line, after an
	// optional list dash.
	yamlKey = regexp.MustCompile(`^(\s*)(-\s+)?([A-Za-z0-9_.:/-]+|"[^"]*"|'[^']*')(\s*:)(\s|$)`)
	// yamlDash is a list item without a key.
	yamlDash = regexp.MustCompile(`^(\s*)(-)(\s|$)`)
	// scalarParts are the parts of a value: interpolations, quoted
	// strings, and words.
	scalarParts = regexp.MustCompile(`\$\{\{.*?\}\}|\$\(\(?[^)]*\)?\)|"(?:[^"\\]|\\.)*"|'[^']*'|\S+`)
	number      = regexp.MustCompile(`^(-?\d+(\.\d+)?|true|false|null|~)$`)
)

// yamlLine marks a line of YAML: the comment, the key, the value.
func yamlLine(line string) string {
	body, comment := splitComment(line)
	var out strings.Builder
	rest := body
	if m := yamlKey.FindStringSubmatch(rest); m != nil {
		out.WriteString(m[1])
		if m[2] != "" {
			out.WriteString(span("p", strings.TrimRight(m[2], " \t")) + m[2][len(strings.TrimRight(m[2], " \t")):])
		}
		out.WriteString(span("k", m[3]) + html.EscapeString(m[4]) + m[5])
		rest = rest[len(m[0]):]
	} else if m := yamlDash.FindStringSubmatch(rest); m != nil {
		out.WriteString(m[1] + span("p", m[2]) + m[3])
		rest = rest[len(m[0]):]
	}
	out.WriteString(scalar(rest))
	if comment != "" {
		out.WriteString(span("c", comment))
	}
	return out.String()
}

// scalar marks the parts of a value: interpolations, strings, numbers.
func scalar(s string) string {
	var out strings.Builder
	last := 0
	for _, loc := range scalarParts.FindAllStringIndex(s, -1) {
		out.WriteString(html.EscapeString(s[last:loc[0]]))
		part := s[loc[0]:loc[1]]
		switch {
		case strings.HasPrefix(part, "${{") || strings.HasPrefix(part, "$("):
			out.WriteString(span("v", part))
		case strings.HasPrefix(part, `"`) || strings.HasPrefix(part, "'"):
			out.WriteString(span("s", part))
		case number.MatchString(part):
			out.WriteString(span("n", part))
		default:
			out.WriteString(markVars(part))
		}
		last = loc[1]
	}
	out.WriteString(html.EscapeString(s[last:]))
	return out.String()
}

// splitComment cuts a trailing # comment off a line, outside quotes.
func splitComment(line string) (string, string) {
	var quote rune
	for i, r := range line {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t'):
			return line[:i], line[i:]
		}
	}
	return line, ""
}

var (
	makeTarget = regexp.MustCompile(`^([A-Za-z0-9_.%/-]+)(\s*::?)(.*)$`)
	makeVar    = regexp.MustCompile(`\$\([^)]*\)|\$\{[^}]*\}|\$[A-Za-z@<^?*]`)
)

// makeLine marks a line of a Makefile: a target, a recipe's variables,
// a comment.
func makeLine(line string) string {
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return span("c", line)
	}
	body, comment := splitComment(line)
	var out string
	if m := makeTarget.FindStringSubmatch(body); m != nil && !strings.HasPrefix(body, "\t") && !strings.Contains(m[3], "=") {
		out = span("k", m[1]) + html.EscapeString(m[2]) + markVars(m[3])
	} else {
		out = markVars(body)
	}
	if comment != "" {
		out += span("c", comment)
	}
	return out
}

// markVars escapes text and marks the $(VAR) references in it.
func markVars(s string) string {
	var out strings.Builder
	last := 0
	for _, loc := range makeVar.FindAllStringIndex(s, -1) {
		out.WriteString(html.EscapeString(s[last:loc[0]]))
		out.WriteString(span("v", s[loc[0]:loc[1]]))
		last = loc[1]
	}
	out.WriteString(html.EscapeString(s[last:]))
	return out.String()
}
