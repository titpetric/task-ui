package config

import (
	"regexp"
	"strings"
)

// context is context.md: what task-ui is and each of its areas, for an
// agent, with the permissions that govern each.
var context = mustRead("context.md")

// mustRead reads an embedded text file; one that is missing is a build
// that cannot run, which the tests catch.
func mustRead(name string) string {
	data, err := files.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Context returns context.md, which the MCP server and every Claude run task-ui starts read.
func Context() string {
	return context
}

// headingKeys reads the keys a section heading names in brackets, as
// "## Inventory (`inventory`, `hosts`)".
var headingKeys = regexp.MustCompile("`([a-z][a-z0-9_-]*)`")

// About returns the first sentence of the context.md section whose heading names key,
// a section or a permission of permissions.yml, or "" when no heading names it.
func About(key string) string {
	for _, block := range strings.Split(context, "\n## ")[1:] {
		heading, body, _ := strings.Cut(block, "\n")
		_, keys, ok := strings.Cut(heading, " (")
		if !ok {
			continue
		}
		for _, m := range headingKeys.FindAllStringSubmatch(keys, -1) {
			if m[1] == key {
				return firstSentence(body)
			}
		}
	}
	return ""
}

// firstSentence is the text of a section up to the end of its first
// sentence, on one line.
func firstSentence(body string) string {
	text := strings.Join(strings.Fields(body), " ")
	if i := strings.Index(text, ". "); i >= 0 {
		return text[:i+1]
	}
	return text
}
