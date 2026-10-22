package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A note is markdown: a mention marked, a line break kept, raw HTML as
// text, and bold, lists and fences as markdown has them.
func TestMessage(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "<p><span class=\"mention\">@claude</span> update <code>weather.php</code><br />\nand push</p>\n", string(view.Message("@claude update `weather.php`\nand push")))
	assert.Equal(t, "<p>a</p>\n\n<p>b &lt;i&gt;</p>\n", string(view.Message("a\n\nb <i>")))
	assert.Equal(t, "<p>Done:</p>\n\n<pre><code class=\"language-text\">ok  pkg\nexit 0\n</code></pre>\n\n<p>bye</p>\n", string(view.Message("Done:\n```text\nok  pkg\nexit 0\n```\nbye")))
	assert.Equal(t, "<p><strong>bold</strong> and <em>soft</em></p>\n\n<ul>\n<li>one</li>\n<li>two <span class=\"mention\">@ops</span></li>\n</ul>\n", string(view.Message("**bold** and *soft*\n\n- one\n- two @ops")))
	assert.Equal(t, "<p>mail me@example.com</p>\n", string(view.Message("mail me@example.com")))
	assert.Equal(t, "<p>run <code>@claude</code> here</p>\n", string(view.Message("run `@claude` here")))
	assert.Equal(t, ``, string(view.Message("")))
}

// A page links a mention of a user to their profile, for a signed in
// user who may open it; a name no user holds stays marked.
func TestPage_Message(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Roles = map[string]string{"ops": model.RoleRunner}
	assert.Equal(t, "<p><a class=\"mention\" href=\"/users/ops\">@ops</a> ask <span class=\"mention\">@nobody</span></p>\n", string(p.Message("@ops ask @nobody")))
	assert.Equal(t, "<p><span class=\"mention\">@ops</span></p>\n", string((&view.Page{}).Message("@ops")))
}
