package view

import (
	"strings"
	"unicode/utf8"

	blackfriday "github.com/russross/blackfriday/v2"
)

// excerptLength is how many characters an excerpt keeps before its
// ellipsis, about two lines of a list row.
const excerptLength = 240

// Excerpt returns a note's markdown as one line of plain text, the words of its headings, emphasis, links and code with the marks gone, cut to excerptLength with an ellipsis.
// A profile lists notes and prompts this way, one a row.
func Excerpt(text string) string {
	var words []string
	node := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions)).Parse([]byte(text))
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && (n.Type == blackfriday.Text || n.Type == blackfriday.Code || n.Type == blackfriday.CodeBlock) {
			words = append(words, strings.Fields(string(n.Literal))...)
		}
		return blackfriday.GoToNext
	})
	line := strings.Join(words, " ")
	if utf8.RuneCountInString(line) <= excerptLength {
		return line
	}
	runes := []rune(line)
	return strings.TrimRight(string(runes[:excerptLength]), " ") + "…"
}
