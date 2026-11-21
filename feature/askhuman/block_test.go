package askhuman_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/feature/askhuman"
	"github.com/titpetric/task-ui/internal/assert"
)

// reply is a note's body: two form blocks, a fence that is no form, a
// longer fence that holds one, and a block Parse refuses.
var reply = strings.Join([]string{
	"A reply with forms.",
	"",
	"```ask",
	"fields:",
	"  - {key: a, component: text}",
	"```",
	"",
	"```yaml",
	"fields: [",
	"```",
	"",
	"````markdown",
	"```ask",
	"fields:",
	"  - {key: never, component: text}",
	"```",
	"````",
	"",
	"- a list item:",
	"",
	"  ```ask",
	"  fields:",
	"    - {key: b, component: text}",
	"  ```",
	"",
	"```ask",
	"fields: [",
	"```",
	"",
}, "\n")

func TestBlocks(t *testing.T) {
	t.Parallel()
	blocks := askhuman.Blocks(reply)
	assert.Len(t, blocks, 3)

	assert.NoError(t, blocks[0].Err)
	assert.Equal(t, "fields:\n  - {key: a, component: text}\n", blocks[0].Body)
	assert.Equal(t, "a", blocks[0].Form.Fields[0].Key)
	assert.Equal(t, "```ask\nfields:\n  - {key: a, component: text}\n```", reply[blocks[0].Start:blocks[0].End])

	assert.NoError(t, blocks[1].Err)
	assert.Equal(t, "fields:\n  - {key: b, component: text}\n", blocks[1].Body)
	assert.Equal(t, "b", blocks[1].Form.Fields[0].Key)
	assert.True(t, strings.HasPrefix(reply[blocks[1].Start:blocks[1].End], "  ```ask\n"))
	assert.True(t, strings.HasSuffix(reply[blocks[1].Start:blocks[1].End], "\n  ```"))

	assert.Nil(t, blocks[2].Form)
	assert.ErrorContains(t, blocks[2].Err, "ask:")

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
		"a backtick inside": {"```ask`yaml\nfields: []\n```\n", 0},
		"another language":  {"```askula\nfields: []\n```\n", 0},
		"an info string":    {"```ask title=one\nfields:\n  - {key: a, component: text}\n```\n", 1},
		"no trailing break": {"```ask\nfields:\n  - {key: a, component: text}\n```", 1},
		"windows breaks":    {"```ask\r\nfields:\r\n  - {key: a, component: text}\r\n```\r\n", 1},
		"a fence of four":   {"````ask\nfields:\n  - {key: a, component: text}\n````\n", 1},
		"never closed":      {"```ask\nfields:\n  - {key: a, component: text}\n", 1},
		"an empty block":    {"```ask\n```\n", 1},
		"two in a row":      {"```ask\nfields:\n  - {key: a, component: text}\n```\n```ask\nfields: []\n```\n", 2},
	} {
		blocks := askhuman.Blocks(test.markdown)
		assert.Len(t, blocks, test.want, name)
		for _, block := range blocks {
			assert.True(t, block.Start < block.End, name)
			assert.True(t, block.End <= len(test.markdown), name)
		}
	}
}

// TestBlockAt reads the ask at an index and names the one the body lacks.
func TestBlockAt(t *testing.T) {
	t.Parallel()
	body := "one:\n\n```ask\nfields:\n  - {key: a, component: text}\n```\n\ntwo:\n\n```ask\nfields: []\n```\n"
	first, err := askhuman.BlockAt(body, 0)
	assert.NoError(t, err)
	assert.Equal(t, "a", first.Fields[0].Key)
	_, err = askhuman.BlockAt(body, 1)
	assert.ErrorContains(t, err, "declares no field")
	_, err = askhuman.BlockAt(body, 2)
	assert.ErrorContains(t, err, "no ask 3")
}

func TestBlock(t *testing.T) {
	t.Parallel()
	blocks := askhuman.Blocks("```ask\nfields:\n  - {key: a, component: text}\nsubmit: {job: \"@shell\"}\n```\n")
	assert.Len(t, blocks, 1)
	assert.Equal(t, askhuman.Block{
		Start: blocks[0].Start,
		End:   blocks[0].End,
		Body:  "fields:\n  - {key: a, component: text}\nsubmit: {job: \"@shell\"}\n",
		Form:  blocks[0].Form,
	}, blocks[0])
	assert.Equal(t, 0, blocks[0].Start)
	assert.Equal(t, "@shell", blocks[0].Form.Submit.Job)
	assert.Equal(t, "Run", blocks[0].Form.Submit.Label)
}
