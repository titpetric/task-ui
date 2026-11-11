package phpfront_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

func TestWithModule(t *testing.T) {
	t.Parallel()
	want := &phpfront.Module{Vendor: "titpetric", Plugin: "refinement", Root: "/tmp/x"}

	got, ok := phpfront.ModuleFrom(phpfront.WithModule(context.Background(), want))
	assert.True(t, ok)
	assert.Equal(t, want, got)
}
