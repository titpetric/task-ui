package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// An integration's view carries the kind, its plain values and the names
// of the secrets that hold a value.
func TestIntegrationView(t *testing.T) {
	t.Parallel()
	view := api.IntegrationView{Kind: *integration.KindOf(integration.Beszel), Values: map[string]string{"url": "u"}, Set: []string{"token"}}
	data, err := json.Marshal(view)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"id":"beszel"`)
	assert.Contains(t, string(data), `"set":["token"]`)
}
