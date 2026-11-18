package search

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokenize returns the lowercased words of a text, two characters or longer.
func Tokenize(text string) []string {
	var tokens []string
	start := -1
	for i, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			addToken(&tokens, text[start:i])
			start = -1
		}
	}
	if start >= 0 {
		addToken(&tokens, text[start:])
	}
	return tokens
}

// addToken appends a lowercased token, dropping one shorter than two characters.
func addToken(tokens *[]string, token string) {
	token = strings.ToLower(token)
	if utf8.RuneCountInString(token) < 2 {
		return
	}
	*tokens = append(*tokens, token)
}

// Terms counts a section's words: a heading outranks its body, a title outranks both.
func Terms(title, heading, body string) map[string]int {
	terms := map[string]int{}
	for _, token := range Tokenize(body) {
		terms[token]++
	}
	for _, token := range Tokenize(heading) {
		terms[token] += 2
	}
	for _, token := range Tokenize(title) {
		terms[token] += 3
	}
	return terms
}

// Unique keeps one of each token, in order.
func Unique(tokens []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if seen[token] {
			continue
		}
		seen[token] = true
		out = append(out, token)
	}
	return out
}

// Snippet returns the text around a matching term, whitespace collapsed.
func Snippet(body string, terms []string, width int) string {
	if width <= 0 {
		width = 160
	}
	runes := []rune(body)
	lower := []rune(strings.ToLower(body))
	at := -1
	for _, term := range terms {
		if term == "" {
			continue
		}
		needle := []rune(strings.ToLower(term))
		for i := 0; i+len(needle) <= len(lower); i++ {
			if string(lower[i:i+len(needle)]) == string(needle) {
				if at < 0 || i < at {
					at = i
				}
				break
			}
		}
	}
	if at < 0 {
		at = 0
	}
	start := max(at-width, 0)
	end := min(at+width, len(runes))
	text := strings.Join(strings.Fields(string(runes[start:end])), " ")
	if start > 0 {
		text = "..." + text
	}
	if end < len(runes) {
		text += "..."
	}
	return text
}
