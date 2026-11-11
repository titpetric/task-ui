package phpfront_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

func TestRegistry(t *testing.T) {
	t.Parallel()
	var registry phpfront.Registry = &fakeRegistry{}
	assert.NotNil(t, registry)

	m := &phpfront.Module{Vendor: "titpetric", Plugin: "refinement", Root: "/x"}
	assert.NoError(t, registry.Register(t.Context(), m))

	dropped, err := registry.Forget(t.Context(), []string{m.Name()})
	assert.NoError(t, err)
	assert.Equal(t, 0, dropped)
	assert.Equal(t, []string{"titpetric/refinement"}, registry.(*fakeRegistry).names)
}
