package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestPromptRequest(t *testing.T) {
	t.Parallel()
	var req api.PromptRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"prompt":"list the skills","cols":120,"rows":40}`), &req))
	assert.Equal(t, api.PromptRequest{Prompt: "list the skills", Cols: 120, Rows: 40}, req)
}
