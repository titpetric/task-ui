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

// A run's view links its report page when it has one, and leaves the
// field out otherwise.
func TestRunView_report(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.RunView{Run: model.Run{ID: "r1"}, Report: "/runs/r1/report"})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"report":"/runs/r1/report"`)
	data, err = json.Marshal(api.RunView{Run: model.Run{ID: "r1"}})
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"report"`)
}
