package form_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront/form"
)

// reply is a note's body: two form blocks, a fence that is no form, a
// longer fence that holds one, and a block Parse refuses.
var reply = strings.Join([]string{
	"A reply with forms.",
	"",
	"```form",
	"fields:",
	"  - {key: a, component: text}",
	"```",
	"",
	"```yaml",
	"fields: [",
	"```",
	"",
	"````markdown",
	"```form",
	"fields:",
	"  - {key: never, component: text}",
	"```",
	"````",
	"",
	"- a list item:",
	"",
	"  ```form",
	"  fields:",
	"    - {key: b, component: text}",
	"  ```",
	"",
	"```form",
	"fields: [",
	"```",
	"",
}, "\n")

func TestBlocks(t *testing.T) {
	t.Parallel()
	blocks := form.Blocks(reply)
	assert.Len(t, blocks, 3)

	assert.NoError(t, blocks[0].Err)
	assert.Equal(t, "fields:\n  - {key: a, component: text}\n", blocks[0].Body)
	assert.Equal(t, "a", blocks[0].Form.Fields[0].Key)
	assert.Equal(t, "```form\nfields:\n  - {key: a, component: text}\n```", reply[blocks[0].Start:blocks[0].End])

	assert.NoError(t, blocks[1].Err)
	assert.Equal(t, "fields:\n  - {key: b, component: text}\n", blocks[1].Body)
	assert.Equal(t, "b", blocks[1].Form.Fields[0].Key)
	assert.True(t, strings.HasPrefix(reply[blocks[1].Start:blocks[1].End], "  ```form\n"))
	assert.True(t, strings.HasSuffix(reply[blocks[1].Start:blocks[1].End], "\n  ```"))

	assert.Nil(t, blocks[2].Form)
	assert.ErrorContains(t, blocks[2].Err, "form:")

	for _, block := range blocks {
		assert.NotContains(t, reply[block.Start:block.End], "never")
	}
}

func TestBlocks_Edges(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		markdown string
		want     int
	}{
		"no fence":          {"a reply\n", 0},
		"no form":           {"```\nfields: []\n```\n", 0},
		"a fence of two":    {"``form\nfields: []\n``\n", 0},
		"a backtick inside": {"```form`yaml\nfields: []\n```\n", 0},
		"another language":  {"```formula\nfields: []\n```\n", 0},
		"an info string":    {"```form title=one\nfields:\n  - {key: a, component: text}\n```\n", 1},
		"no trailing break": {"```form\nfields:\n  - {key: a, component: text}\n```", 1},
		"windows breaks":    {"```form\r\nfields:\r\n  - {key: a, component: text}\r\n```\r\n", 1},
		"a fence of four":   {"````form\nfields:\n  - {key: a, component: text}\n````\n", 1},
		"never closed":      {"```form\nfields:\n  - {key: a, component: text}\n", 1},
		"an empty block":    {"```form\n```\n", 1},
		"two in a row":      {"```form\nfields:\n  - {key: a, component: text}\n```\n```form\nfields: []\n```\n", 2},
	} {
		blocks := form.Blocks(test.markdown)
		assert.Len(t, blocks, test.want, name)
		for _, block := range blocks {
			assert.True(t, block.Start < block.End, name)
			assert.True(t, block.End <= len(test.markdown), name)
		}
	}
}

func TestBlock(t *testing.T) {
	t.Parallel()
	blocks := form.Blocks("```form\nfields:\n  - {key: a, component: text}\nsubmit: {job: \"@shell\"}\n```\n")
	assert.Len(t, blocks, 1)
	assert.Equal(t, form.Block{
		Start: blocks[0].Start,
		End:   blocks[0].End,
		Body:  "fields:\n  - {key: a, component: text}\nsubmit: {job: \"@shell\"}\n",
		Form:  blocks[0].Form,
	}, blocks[0])
	assert.Equal(t, 0, blocks[0].Start)
	assert.Equal(t, "@shell", blocks[0].Form.Submit.Job)
	assert.Equal(t, "Run", blocks[0].Form.Submit.Label)
}
