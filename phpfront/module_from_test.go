package phpfront_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

func TestModuleFrom(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, ok := phpfront.ModuleFrom(ctx)
	assert.False(t, ok, "a bare context names no module")

	_, ok = phpfront.ModuleFrom(phpfront.WithModule(ctx, nil))
	assert.False(t, ok, "a nil module is no module")

	m := &phpfront.Module{Vendor: "acme", Plugin: "notes", Root: "/tmp/y"}
	got, ok := phpfront.ModuleFrom(phpfront.WithModule(ctx, m))
	assert.True(t, ok)
	assert.Equal(t, "acme/notes", got.Name())
}
