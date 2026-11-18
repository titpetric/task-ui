package search_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
)

// Words are lowercased and one character is not a term.
func TestTokenize(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"ci", "daily", "runs", "at", "0300"}, search.Tokenize("CI:daily runs at 0300, x"))
}

// A heading outranks its body; a title outranks both.
func TestTerms(t *testing.T) {
	t.Parallel()
	terms := search.Terms("Run book", "Install", "install the server")
	assert.Equal(t, 3, terms["install"])
	assert.Equal(t, 1, terms["server"])
	assert.Equal(t, 3, terms["run"])
}

// A token is kept once, in the order it arrived.
func TestUnique(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a1", "b2"}, search.Unique([]string{"a1", "b2", "a1"}))
}

// A snippet is a window on the match with its whitespace collapsed.
func TestSnippet(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("filler ", 40) + "the daily deploy runs here " + strings.Repeat("tail ", 40)
	snippet := search.Snippet(body, []string{"daily"}, 20)
	assert.Contains(t, snippet, "daily deploy")
	assert.True(t, strings.HasPrefix(snippet, "..."))
	assert.True(t, strings.HasSuffix(snippet, "..."))
	assert.False(t, strings.Contains(snippet, "  "))

	assert.Equal(t, "short text", search.Snippet("short   text", nil, 80))
}
