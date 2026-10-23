package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A summary renders as markdown: headings, lists, code and links, with
// raw HTML and scripts shown as text, never markup.
func TestMarkdown(t *testing.T) {
	t.Parallel()
	out := string(view.Markdown("## Build\n\n- image `api:1.2`\n- pushed\n\n```\nok\n```\n\n[log](https://example.com/log) <script>alert(1)</script> [x](javascript:alert(1))"))
	assert.Contains(t, out, "<h2")
	assert.Contains(t, out, "<li>image <code>api:1.2</code></li>")
	assert.Contains(t, out, "<pre><code>ok\n</code></pre>")
	assert.Contains(t, out, `href="https://example.com/log"`)
	assert.Contains(t, out, "&lt;script&gt;")
	assert.NotContains(t, out, "<script>")
	assert.NotContains(t, out, `href="javascript:`)
}
