package phpscript_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestModuleFrom(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, ok := phpscript.ModuleFrom(ctx)
	assert.False(t, ok, "a bare context names no module")

	_, ok = phpscript.ModuleFrom(phpscript.WithModule(ctx, nil))
	assert.False(t, ok, "a nil module is no module")

	m := &phpscript.Module{Vendor: "acme", Plugin: "notes", Root: "/tmp/y"}
	got, ok := phpscript.ModuleFrom(phpscript.WithModule(ctx, m))
	assert.True(t, ok)
	assert.Equal(t, "acme/notes", got.Name())
}
