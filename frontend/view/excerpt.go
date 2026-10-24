package view

import (
	"strings"
	"unicode/utf8"
)

// excerptMarks are the marks that start a line of markdown: a heading,
// a quote, a list's bullet.
const excerptMarks = "#>*-+ \t"

// Excerpt is markdown as a list shows it: its words on one line, without marks, fences or tables;
// past n runes, cut at a word with an ellipsis.
func Excerpt(text string, n int) string {
	var words []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") || strings.HasPrefix(line, "|") {
			continue
		}
		line = strings.TrimLeft(line, excerptMarks)
		line = strings.NewReplacer("**", "", "__", "", "`", "").Replace(line)
		words = append(words, strings.Fields(line)...)
	}
	short := strings.Join(words, " ")
	if utf8.RuneCountInString(short) <= n {
		return short
	}
	cut := string([]rune(short)[:n])
	if space := strings.LastIndex(cut, " "); space > n/2 {
		cut = cut[:space]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}
