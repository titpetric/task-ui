package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A reply's form blocks come back parsed, in order, and a block the
// dialect refuses comes back as the reason: a client draws the question
// or the error without reading yaml itself.
func TestForms_List(t *testing.T) {
	t.Parallel()
	h := api.New(api.Options{}).Routes()
	admin := &model.User{ID: "1", Name: "root", Admin: true, Role: model.RoleRunner}
	body := "Shall I?\n\n```form\n" +
		"fields:\n  - key: cache\n    label: Cache the list page?\n    component: bool\n" +
		"submit:\n  label: Answer\n  prompt: \"cache: {{cache}}\"\n" +
		"```\n\nand one that is wrong:\n\n```form\nfields: []\n```\n"
	rec := voiceDo(h, admin, http.MethodPost, "/forms", body)
	assert.Equal(t, http.StatusOK, rec.Code)

	var out struct {
		Forms []struct {
			Form *struct {
				Fields []struct {
					Key       string `json:"key"`
					Label     string `json:"label"`
					Component string `json:"component"`
				} `json:"fields"`
				Submit struct {
					Label  string `json:"label"`
					Prompt string `json:"prompt"`
				} `json:"submit"`
			} `json:"form"`
			Error string `json:"error"`
		} `json:"forms"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Len(t, out.Forms, 2)
	assert.NotNil(t, out.Forms[0].Form)
	assert.Equal(t, "", out.Forms[0].Error)
	assert.Len(t, out.Forms[0].Form.Fields, 1)
	assert.Equal(t, "cache", out.Forms[0].Form.Fields[0].Key)
	assert.Equal(t, "Cache the list page?", out.Forms[0].Form.Fields[0].Label)
	assert.Equal(t, "bool", out.Forms[0].Form.Fields[0].Component)
	assert.Equal(t, "Answer", out.Forms[0].Form.Submit.Label)
	assert.Equal(t, "cache: {{cache}}", out.Forms[0].Form.Submit.Prompt)
	assert.Nil(t, out.Forms[1].Form)
	assert.Contains(t, out.Forms[1].Error, "no field")

	// A body with no fence holds no form, and a reader who may not read a
	// prompt may not read its forms either.
	rec = voiceDo(h, admin, http.MethodPost, "/forms", "nothing here")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"forms":[]`)

	viewer := &model.User{ID: "2", Name: "view", Role: model.RoleViewer}
	rec = voiceDo(h, viewer, http.MethodPost, "/forms", body)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "prompt/view")
}
