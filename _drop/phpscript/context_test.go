package phpscript_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestWithModule(t *testing.T) {
	t.Parallel()
	want := &phpscript.Module{Vendor: "titpetric", Plugin: "refinement", Root: "/tmp/x"}

	got, ok := phpscript.ModuleFrom(phpscript.WithModule(context.Background(), want))
	assert.True(t, ok)
	assert.Equal(t, want, got)
}
