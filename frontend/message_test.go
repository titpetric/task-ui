package phpfront

import (
	"testing"

	"github.com/titpetric/phpscript/model"

	"github.com/titpetric/task-ui/internal/assert"
)

// A note's text is rendered through the renderer that owns the content
// security policy, so raw HTML in it shows as text and never as markup. The
// mentions are marked on the escaped text, after that.
func TestMessage(t *testing.T) {
	t.Parallel()

	out := Message("see **this** and <script>alert(1)</script>", "/users/", nil)
	assert.Contains(t, out, "<strong>this</strong>")
	assert.NotContains(t, out, "<script>", "the renderer escapes what the author wrote")
	assert.Contains(t, out, "&lt;script&gt;")
}

// A line break is kept as one, because a note is written in a textarea and
// read back as it was typed.
func TestMessage_HardBreaks(t *testing.T) {
	t.Parallel()

	assert.Contains(t, Message("one\ntwo", "/users/", nil), "<br")
}

// A name the instance knows becomes a link under the prefix, which carries the
// mount path; every other mention is marked and not linked, so a note naming
// somebody who left still reads as naming them.
func TestMessage_Mentions(t *testing.T) {
	t.Parallel()

	out := Message("@ann and @bob", "/development/users/", []string{"ann"})
	assert.Contains(t, out, `<a class="mention" href="/development/users/ann">@ann</a>`)
	assert.Contains(t, out, `<span class="mention">@bob</span>`)

	// A name needing escaping in a path is escaped in the href and left as
	// written in the text.
	escaped := Message("@a b", "/users/", []string{"a"})
	assert.Contains(t, escaped, `href="/users/a"`)
}

// A mention is only a mention where a note names a user. A job ID reads as
// @kind:name, and the colon is what tells the two apart: marking one would
// link a terminal to a user page that does not exist.
func TestMessage_JobIsNoMention(t *testing.T) {
	t.Parallel()

	out := Message("ran @host:dev1 and @shell", "/users/", []string{"host", "shell"})
	assert.Contains(t, out, "@host:dev1")
	assert.NotContains(t, out, `href="/users/host"`)
	assert.Contains(t, out, `<a class="mention" href="/users/shell">@shell</a>`, "no colon is a name")
}

// What leads a mention is kept: a name after an opening bracket or a comma is
// marked, and one inside a longer word is not, which is how an email address
// in a note survives.
func TestMessage_Leads(t *testing.T) {
	t.Parallel()

	out := Message("(@ann, @bob)", "/users/", []string{"ann", "bob"})
	assert.Contains(t, out, "(<a class=\"mention\"")
	assert.Contains(t, out, ", <a class=\"mention\"")

	assert.NotContains(t, Message("mail ops@example.com", "/users/", []string{"example"}), "mention")
}

// names reads whatever spelling of a list the runtime hands over, because a
// page asks about many rows at once with array_keys and that answers a Go
// slice rather than an array.
func TestNames(t *testing.T) {
	t.Parallel()

	array := model.NewArraySize(2)
	array.Append("infra")
	array.Append("malta/deploy")
	assert.Equal(t, []string{"infra", "malta/deploy"}, names(array), "an array a script wrote")

	assert.Equal(t, []string{"infra"}, names([]any{"infra"}), "what array_keys answers")
	assert.Equal(t, []string{"infra"}, names([]string{"infra"}))

	// A blank entry is no name, so a list with a hole in it asks about the
	// rows that are there.
	holed := model.NewArraySize(3)
	holed.Append("infra")
	holed.Append("  ")
	holed.Append(nil)
	assert.Equal(t, []string{"infra"}, names(holed))

	// A value that is no list reads as no names, rather than panicking in a
	// binding a template called.
	assert.Equal(t, []string{}, names(nil))
	assert.Equal(t, []string{}, names("infra"))
	assert.Equal(t, []string{}, names(42))
}

// fields keys the way PHP keys: a string key as itself, and a list index as
// its decimal, so a form posted as a list and one posted as a map both reach
// the Go caller as a map.
func TestFields(t *testing.T) {
	t.Parallel()

	array := model.NewArraySize(2)
	array.Set("token", "t")
	array.Set("who", "ops")
	assert.Equal(t, map[string]any{"token": "t", "who": "ops"}, fields(array))

	list := model.NewArraySize(2)
	list.Append("first")
	list.Append("second")
	assert.Equal(t, map[string]any{"0": "first", "1": "second"}, fields(list))

	assert.Equal(t, map[string]any{}, fields(nil))
}
