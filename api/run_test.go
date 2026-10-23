package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A run's view carries the title its terminal set.
func TestRunView_title(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.RunView{Run: model.Run{ID: "r1", Title: "deploying"}})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"title":"deploying"`)
}
