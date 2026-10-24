package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// An excerpt is a note's text on one line with the markdown gone: the
// words of a heading, emphasis, a link and code stay, the marks and the
// link's address go, and a long note ends in an ellipsis.
func TestExcerpt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bold # head l", view.Excerpt("**bold** # head [l](u)"), "a # inside a line is text")
	assert.Equal(t, "head body", view.Excerpt("# head\n\nbody"))
	assert.Equal(t, "rotate the certificates", view.Excerpt("rotate the certificates"))
	assert.Equal(t, "run make build then deploy", view.Excerpt("run `make build`\n\n- then\n- deploy\n"))
	assert.Equal(t, "", view.Excerpt(""))
	long := view.Excerpt(strings.Repeat("word ", 100))
	assert.True(t, strings.HasSuffix(long, "…"))
	assert.True(t, len([]rune(long)) <= 241)
}
