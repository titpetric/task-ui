package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestMessage(t *testing.T) {
	t.Parallel()
	assert.Equal(t, `<p><span class="mention">@claude</span> update <code>weather.php</code><br>and push</p>`, string(view.Message("@claude update `weather.php`\nand push")))
	assert.Equal(t, `<p>a</p><p>b &lt;i&gt;</p>`, string(view.Message("a\n\nb <i>")))
	assert.Equal(t, `<p>Done:</p><pre class="note-code">ok  pkg
exit 0</pre><p>bye</p>`, string(view.Message("Done:\n```text\nok  pkg\nexit 0\n```\nbye")))
	assert.Equal(t, "<p>open ``` fence</p>", string(view.Message("open ``` fence")))
	assert.Equal(t, `<p>mail me@example.com</p>`, string(view.Message("mail me@example.com")))
	assert.Equal(t, ``, string(view.Message("")))
}
