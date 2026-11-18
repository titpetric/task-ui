package app

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// An instance with no setting table, as a unit test has none, is still named:
// the identifier is made and answered, just not kept.
func TestInstanceIDNoSettings(t *testing.T) {
	t.Parallel()

	id, err := instanceID(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, model.ValidID(id))
}
