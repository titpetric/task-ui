package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The listing has the folders a user sees, each job with whether they may
// run it and its latest run.
func TestFolder_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.finished(t, f.start(t, asOps, "infra", "plan"))

	decode := func(user, query string) []api.FolderView {
		res := f.call(t, user, http.MethodGet, "/folders"+query, nil)
		assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
		var body struct {
			Folders []api.FolderView `json:"folders"`
		}
		assert.NoError(t, json.Unmarshal(res.Raw, &body))
		return body.Folders
	}
	ids := func(folders []api.FolderView) []string {
		out := []string{}
		for _, folder := range folders {
			out = append(out, folder.ID)
		}
		return out
	}

	assert.Equal(t, []string{"infra", "infra/edge", "private"}, ids(decode(asAdmin, "")))
	assert.Equal(t, []string{"infra"}, ids(decode(asLimited, "")))

	infra := decode(asOps, "?folder=infra")
	assert.Len(t, infra, 1)
	var plan api.JobView
	for _, job := range infra[0].Jobs {
		if job.ID == "plan" {
			plan = job
		}
	}
	assert.True(t, plan.CanRun)
	assert.NotNil(t, plan.LastRun)
	assert.Equal(t, model.StateOK, plan.LastRun.State)

	for _, job := range decode(asViewer, "?folder=infra")[0].Jobs {
		assert.False(t, job.CanRun, job.ID)
	}
}
