// Package search indexes documentation and job targets for the search page and
// the MCP tools that answer from them.
package search

import (
	"path"
	"strings"

	blackfriday "github.com/russross/blackfriday/v2"
)

// Section holds a heading with the text under it, in the document's order.
type Section struct {
	Ordinal    int
	Level      int
	Heading    string
	Breadcrumb string
	Line       int
	Body       string
}

// extensions are the markdown features the blog renderer reads, blackfriday's common set.
const extensions = blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.NoEmptyLineBeforeBlock

// Split walks the blackfriday AST: a heading starts a section, a fenced block
// keeps its text, and front matter is skipped. The title is the first H1, else
// name without its extension.
func Split(src []byte, name string) (string, []Section) {
	lines := strings.Split(normalize(src), "\n")
	begin := frontMatter(lines)
	root := blackfriday.New(blackfriday.WithExtensions(extensions)).Parse([]byte(strings.Join(lines[begin:], "\n")))

	var (
		sections          []Section
		stack             [7]string
		body              []string
		cursor            = begin
		title             string
		currentLevel      int
		currentHeading    string
		currentBreadcrumb string
		currentLine       = firstLine(lines, begin)
	)
	close := func() {
		text := strings.TrimSpace(strings.Join(body, "\n"))
		if currentHeading != "" || text != "" {
			sections = append(sections, Section{
				Ordinal:    len(sections),
				Level:      currentLevel,
				Heading:    currentHeading,
				Breadcrumb: currentBreadcrumb,
				Line:       currentLine,
				Body:       text,
			})
		}
		body = nil
	}
	open := func(level int, heading string, line int) {
		close()
		currentLevel, currentHeading, currentLine = level, heading, line
		currentBreadcrumb = breadcrumb(stack[:], level, heading)
	}
	// The AST carries no positions, so a heading's line is the next source line
	// that names it, read forward from the last one found.
	for node := root.FirstChild; node != nil; node = node.Next {
		if node.Type == blackfriday.Heading {
			heading := nodeText(node)
			if node.Level == 1 && title == "" {
				title = heading
			}
			line := locate(lines, heading, cursor)
			if line > 0 {
				cursor = line - 1
			}
			open(int(node.Level), heading, line)
			continue
		}
		if text := blockText(node); text != "" {
			body = append(body, text)
		}
	}
	close()

	if title == "" {
		title = stem(name)
	}
	if len(sections) == 1 && sections[0].Heading == "" {
		sections[0].Heading, sections[0].Level, sections[0].Breadcrumb = title, 1, title
	}
	return title, sections
}

// SplitText reads a plain text file as one section named by its file.
func SplitText(src []byte, name string) (string, []Section) {
	title := stem(name)
	body := strings.TrimSpace(normalize(src))
	if body == "" {
		return title, nil
	}
	return title, []Section{{Ordinal: 0, Level: 1, Heading: title, Breadcrumb: title, Line: 1, Body: body}}
}

// normalize folds CRLF and lone CR into LF, as the blog renderer does before blackfriday reads a line.
func normalize(src []byte) string {
	content := strings.ReplaceAll(string(src), "\r\n", "\n")
	return strings.ReplaceAll(content, "\r", "\n")
}

// frontMatter returns the first line after a leading `---` block, or zero.
func frontMatter(lines []string) int {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		switch strings.TrimSpace(lines[i]) {
		case "---", "...":
			return i + 1
		}
	}
	return 0
}

// nodeText is the text inside a node, its markup dropped.
func nodeText(node *blackfriday.Node) string {
	var out strings.Builder
	var walk func(*blackfriday.Node)
	walk = func(node *blackfriday.Node) {
		switch node.Type {
		case blackfriday.Text, blackfriday.Code, blackfriday.HTMLBlock:
			out.Write(node.Literal)
		case blackfriday.Softbreak, blackfriday.Hardbreak:
			out.WriteByte('\n')
		}
		for child := node.FirstChild; child != nil; child = child.Next {
			walk(child)
		}
	}
	walk(node)
	return strings.TrimSpace(out.String())
}

// blockText is the text of one block, its blocks kept apart.
func blockText(node *blackfriday.Node) string {
	switch node.Type {
	case blackfriday.CodeBlock:
		return strings.TrimSpace(string(node.Literal))
	case blackfriday.List:
		var items []string
		for item := node.FirstChild; item != nil; item = item.Next {
			var parts []string
			for child := item.FirstChild; child != nil; child = child.Next {
				if text := blockText(child); text != "" {
					parts = append(parts, text)
				}
			}
			if len(parts) > 0 {
				items = append(items, "- "+strings.Join(parts, "\n  "))
			}
		}
		return strings.Join(items, "\n")
	case blackfriday.BlockQuote:
		return blockChildren(node)
	case blackfriday.Table:
		return blockChildren(node)
	default:
		return nodeText(node)
	}
}

// blockChildren joins the text of a container's blocks.
func blockChildren(node *blackfriday.Node) string {
	var parts []string
	for child := node.FirstChild; child != nil; child = child.Next {
		if text := blockText(child); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

// firstLine returns the 1-based number of the first non-blank line from start.
func firstLine(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return i + 1
		}
	}
	return start + 1
}

// locate returns the line naming a heading, read forward from from.
func locate(lines []string, heading string, from int) int {
	needle := strings.TrimSpace(heading)
	if needle == "" {
		return 0
	}
	for i := from; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == needle {
			return i + 1
		}
	}
	for i := from; i < len(lines); i++ {
		if strings.Contains(lines[i], needle) {
			return i + 1
		}
	}
	return 0
}

// breadcrumb is the heading's place in the document, its ancestors then itself.
func breadcrumb(stack []string, level int, heading string) string {
	if level == 0 {
		return ""
	}
	stack[level] = heading
	for i := level + 1; i < len(stack); i++ {
		stack[i] = ""
	}
	parts := make([]string, 0, level)
	for i := 1; i <= level; i++ {
		if stack[i] != "" {
			parts = append(parts, stack[i])
		}
	}
	return strings.Join(parts, " > ")
}

// stem is a file name without its extension.
func stem(name string) string {
	base := path.Base(strings.TrimSpace(name))
	if base == "." || base == "/" || base == "" {
		return name
	}
	return strings.TrimSuffix(base, path.Ext(base))
}
