package phpscript_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestParseName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/m", phpscript.Prefix)

	for _, name := range []string{"titpetric/mysql", "titpetric/refinement", "a-b/c_d.e", "x/y", " titpetric/mysql "} {
		vendor, plugin, err := phpscript.ParseName(name)
		assert.NoError(t, err, name)
		assert.NotEqual(t, "", vendor, name)
		assert.NotEqual(t, "", plugin, name)
	}

	vendor, plugin, err := phpscript.ParseName("titpetric/refinement")
	assert.NoError(t, err)
	assert.Equal(t, "titpetric", vendor)
	assert.Equal(t, "refinement", plugin)

	for _, bad := range []string{"", "titpetric", "/mysql", "titpetric/", "a/b/c", "Titpetric/mysql", "titpetric/My SQL", "../x", "titpetric/..", "./x", "t/x!"} {
		_, _, err := phpscript.ParseName(bad)
		assert.ErrorIs(t, err, phpscript.ErrName, bad)
	}
}
