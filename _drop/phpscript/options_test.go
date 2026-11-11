package phpscript_test

import (
	"errors"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestOptions(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 30*time.Second, phpscript.DefaultTimeout)

	var reported []error
	opts := phpscript.NewOptions("/data/modules")
	opts.OnError = func(err error) { reported = append(reported, err) }
	opts.Roots = append(opts.Roots, "/data/workspaces")
	assert.Equal(t, []string{"/data/modules", "/data/workspaces"}, opts.Roots)

	// A Manager built with nothing set takes every default, so a caller sets
	// the fields its instance needs and leaves the rest.
	m := phpscript.New(phpscript.Options{OnError: opts.OnError})
	assert.Equal(t, phpscript.DefaultTimeout, m.Timeout())
	assert.Len(t, reported, 0)

	opts.OnError(errors.New("a module threw"))
	assert.Len(t, reported, 1)
}
