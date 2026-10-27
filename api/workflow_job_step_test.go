package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// stepFixture returns the workflow routes over a folder the listings
// show, infra, and one the steps' catalogue alone lists, infra/hidden.
func stepFixture(steps bool) http.Handler {
	plan := &folder.Job{Folder: "infra", ID: "plan"}
	plan.Name = "plan"
	infra := &folder.Folder{ID: "infra", Jobs: []*folder.Job{plan}}
	deploy := &folder.Job{Folder: "infra/hidden", ID: "deploy"}
	deploy.Name = "deploy"
	hidden := &folder.Folder{ID: "infra/hidden", Jobs: []*folder.Job{deploy}}
	opts := api.Options{Catalogue: catalogue{"infra": infra}}
	if steps {
		opts.Steps = catalogue{"infra": infra, "infra/hidden": hidden}
	}
	return api.New(opts).Routes()
}

// A job step resolves its folder through the steps' catalogue, every
// folder with a job, as the editor lists them: a folder the listings
// hide has its job checked, not refused as missing. For a folder in
// neither catalogue, or outside the user's own, the error names the step
// by its folder relative to the workspace and the job, and the workspace
// that lacks it.
// Admins, as the routes decide permissions from the user alone here.
func TestWorkflow_JobStep(t *testing.T) {
	t.Parallel()
	ops := &model.User{Name: "ops", Role: model.RoleAdmin}
	limited := &model.User{Name: "limited", Role: model.RoleAdmin, Folders: model.EncodeList([]string{"infra"})}
	h := stepFixture(true)

	code, out := pinCall(t, h, ops, http.MethodPost, "/workflows", `{"name":"x","steps":[{"folder":"infra/hidden","job":"nope"}]}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "step 1: infra/hidden / nope: no such job", out["error"], "the hidden folder is there, its job is not")

	for body, want := range map[string]string{
		`{"name":"x","steps":[{"folder":"nowhere","job":"plan"}]}`:                                        "step 1: nowhere / plan: Internal has no folder nowhere",
		`{"name":"x","steps":[{"folder":"infra","job":"plan"},{"folder":"internal","job":"plan"}]}`:       "step 2: the root / plan: Internal has no jobs at its root",
		`{"name":"x","steps":[{"folder":"infra/hidden","job":"deploy"},{"folder":"internal","job":"x"}]}`: "step 2: the root / x: Internal has no jobs at its root",
	} {
		code, out = pinCall(t, h, ops, http.MethodPost, "/workflows", body)
		assert.Equal(t, http.StatusBadRequest, code, body)
		assert.Equal(t, want, out["error"], body)
	}

	code, out = pinCall(t, h, limited, http.MethodPost, "/workflows", `{"name":"x","steps":[{"folder":"infra/hidden","job":"deploy"}]}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "step 1: infra/hidden / deploy: Internal has no folder infra/hidden", out["error"], "a folder outside the user's own")

	// Without a steps' catalogue the listings decide.
	code, out = pinCall(t, stepFixture(false), ops, http.MethodPost, "/workflows", `{"name":"x","steps":[{"folder":"infra/hidden","job":"deploy"}]}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "step 1: infra/hidden / deploy: Internal has no folder infra/hidden", out["error"])
}
