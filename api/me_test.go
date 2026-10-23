package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// An area marshals its rules only when it has any.
func TestAreaView(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(api.AreaView{Key: "secret", Title: "Secrets"})
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"secret","title":"Secrets","about":"","allowed":false}`, string(raw))
}
