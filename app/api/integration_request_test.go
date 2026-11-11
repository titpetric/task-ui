package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestIntegrationRequest(t *testing.T) {
	t.Parallel()
	var req api.IntegrationRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"values":{"url":"u"}}`), &req))
	assert.Equal(t, "u", req.Values["url"])
}
