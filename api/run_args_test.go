package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestRunArgs_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var req api.RunRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"folder":"malta","dir":"dockerfiles","job":"custom","args":"ps -a '\"; id\"'"}`), &req))
	assert.Equal(t, "dockerfiles", req.Dir)
	assert.Equal(t, api.RunArgs{"ps", "-a", `"; id"`}, req.Args)

	var list api.RunArgs
	assert.NoError(t, json.Unmarshal([]byte(`["logs", "-f", "a b"]`), &list))
	assert.Equal(t, api.RunArgs{"logs", "-f", "a b"}, list)
	assert.NoError(t, json.Unmarshal([]byte(`null`), &list))

	for _, bad := range []string{`"'open"`, `["a\nb"]`, `42`, `[1]`} {
		var args api.RunArgs
		assert.Error(t, json.Unmarshal([]byte(bad), &args), bad)
	}
}
