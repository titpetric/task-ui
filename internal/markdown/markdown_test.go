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

// An image from elsewhere shows as a link to it, under its alt text or
// its address, and never as an image; one of the tree, an absolute
// path or a data URL stays an image. The policy the pages carry allows
// those alone.
func TestRender_ExternalImage(t *testing.T) {
	t.Parallel()
	html := markdown.HTML([]byte("![pixel](https://x.example/px.png) ![](//x.example/px.png) ![js](javascript:alert(1))"), "", "")
	assert.NotContains(t, html, "<img")
	assert.Contains(t, html, `<a href="https://x.example/px.png" target="_blank" rel="nofollow noreferrer">pixel</a>`)
	assert.Contains(t, html, "<tt>//x.example/px.png</tt>", "a link without a scheme is no safe link: its address shows as text")
	assert.Contains(t, html, "<tt>js</tt>")
	assert.NotContains(t, html, "javascript:")

	html = markdown.HTML([]byte("![plot](plot.png) ![icon](/static/icon.svg) ![dot](data:image/gif;base64,R0lGODlhAQABAAAAACw=)"), "/artifacts/files/", "reports")
	assert.Contains(t, html, `<img src="/artifacts/files/reports/plot.png" alt="plot" />`)
	assert.Contains(t, html, `<img src="/static/icon.svg" alt="icon" />`)
	assert.Contains(t, html, `<img src="data:image/gif;base64,R0lGODlhAQABAAAAACw=" alt="dot" />`)

	assert.Contains(t, markdown.CSP, "img-src 'self' data:")
	assert.Contains(t, markdown.CSP, "script-src 'self'")
	assert.NotContains(t, markdown.CSP, "script-src 'self' 'unsafe-inline'")
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

// A list right under a line of text is a list, as a chat writes it; in a
// fence it stays code.
func TestRender_ListAfterText(t *testing.T) {
	t.Parallel()
	html := markdown.Render(markdown.Options{HardBreaks: true}, []byte("**Runtime**\n- one\n- two\n\n```\nx\n- y\n```"))
	assert.Equal(t, "<p><strong>Runtime</strong></p>\n\n<ul>\n<li>one</li>\n<li>two</li>\n</ul>\n\n<pre><code>x\n- y\n</code></pre>\n", html)
}
