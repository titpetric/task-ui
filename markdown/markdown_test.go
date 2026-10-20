package markdown_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/markdown"
)

func TestHTML(t *testing.T) {
	t.Parallel()
	html := markdown.HTML([]byte("# Report\r\n\r\n| a | b |\n|---|---|\n| 1 | 2 |\n\n![plot](plot.png) [up](../x.md) [web](https://example.com) [js](javascript:alert(1))\n\n<script>alert(1)</script>\n"), "/artifacts/files/", "reports")
	assert.Contains(t, html, `<h1 id="report">Report</h1>`)
	assert.Contains(t, html, "<td>1</td>")
	assert.Contains(t, html, `<img src="/artifacts/files/reports/plot.png" alt="plot" />`)
	assert.Contains(t, html, `href="/artifacts/files/x.md"`)
	assert.Contains(t, html, `href="https://example.com" target="_blank" rel="nofollow noreferrer"`)
	assert.NotContains(t, html, "javascript:")
	assert.NotContains(t, html, "<script>")
	assert.Contains(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
	assert.Contains(t, markdown.HTML([]byte("[out](../../../etc/passwd)"), "/f/", "a"), `href="/f/etc/passwd"`)
}

// Render keeps a single line break with HardBreaks, and the hook rewrites
// the text outside code and links.
func TestRender(t *testing.T) {
	t.Parallel()
	mark := func(text string) string { return strings.ReplaceAll(text, "@cl", "<b>@cl</b>") }
	html := markdown.Render(markdown.Options{HardBreaks: true, Text: mark}, []byte("hi @cl\nnext `@cl` [@cl](https://x.example)"))
	assert.Equal(t, "<p>hi <b>@cl</b><br />\nnext <code>@cl</code> <a href=\"https://x.example\" target=\"_blank\" rel=\"nofollow noreferrer\">@cl</a></p>\n", html)
	assert.Contains(t, markdown.Render(markdown.Options{Text: mark}, []byte("a & <b>")), "a &amp; ")
}
