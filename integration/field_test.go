package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// A field marks a secret, and leaves out what it lacks.
func TestField(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(integration.Field{Name: "token", Label: "Token", Secret: true})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"token","label":"Token","secret":true}`, string(data))
}
