package main

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

func TestRun(t *testing.T) {
	t.Parallel()
	var out, errs bytes.Buffer
	assert.Equal(t, 0, run([]string{"../../.."}, &out, &errs), out.String())
	assert.Equal(t, "", out.String())

	assert.Equal(t, 1, run([]string{"../../rulelint/testdata/app"}, &out, &errs))
	assert.Contains(t, out.String(), `unknown rule "note/write"`)
	assert.Contains(t, out.String(), `config: rule "workspace/view" never checked`)

	assert.Equal(t, 2, run([]string{"testdata/none"}, &out, &errs))
	assert.Contains(t, errs.String(), "rulelint:")
}
