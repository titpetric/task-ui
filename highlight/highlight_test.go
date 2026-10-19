package highlight_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/highlight"
	"github.com/titpetric/task-ui/internal/assert"
)

// YAML is drawn with its keys, comments, strings, numbers, list dashes and
// ${{ }} interpolations marked; everything is escaped.
func TestYAML(t *testing.T) {
	t.Parallel()
	src := "# the setup\njobs:\n  setup:\n    cmds:\n      - for: repositories as repo\n        cmd: git clone ${{repo.remote}} \"<x>\"\n    retries: 3\n    internal: true\n"
	out := string(highlight.HTML("atkins.yml", src))

	assert.Contains(t, out, `<span class="c"># the setup</span>`)
	assert.Contains(t, out, `<span class="k">jobs</span>:`)
	assert.Contains(t, out, `  <span class="k">setup</span>:`)
	assert.Contains(t, out, `<span class="p">-</span> <span class="k">for</span>:`)
	assert.Contains(t, out, `<span class="v">${{repo.remote}}</span>`)
	assert.Contains(t, out, `<span class="s">&#34;&lt;x&gt;&#34;</span>`)
	assert.Contains(t, out, `<span class="n">3</span>`)
	assert.Contains(t, out, `<span class="n">true</span>`)
	assert.False(t, strings.Contains(out, "<x>"), "escaped")
	assert.Equal(t, strings.Count(src, "\n"), strings.Count(out, "\n"))
}

// A Makefile marks its targets, comments and variables.
func TestMakefile(t *testing.T) {
	t.Parallel()
	out := string(highlight.HTML("Makefile", "# build it\nbuild: deps ## Compile\n\tgo build -o $(BIN) .\n"))
	assert.Contains(t, out, `<span class="c"># build it</span>`)
	assert.Contains(t, out, `<span class="k">build</span>: deps <span class="c">## Compile</span>`)
	assert.Contains(t, out, `<span class="v">$(BIN)</span>`)
}

// Anything else is escaped and drawn as it is.
func TestPlain(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "echo &lt;hi&gt;\n", string(highlight.HTML("run.sh", "echo <hi>\n")))
}
