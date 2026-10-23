package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// meOf reads GET /me as a user.
func meOf(t *testing.T, f *fixture, user string) api.MeView {
	t.Helper()
	res := f.call(t, user, http.MethodGet, "/me", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.MeView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	return view
}

// /me says what the user may do: the paths, each area with its purpose,
// and the folders the user sees.
func TestMeView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ops := meOf(t, f, asOps)
	assert.Equal(t, "ops", ops.Name)
	assert.Equal(t, "runner", ops.Role)
	assert.False(t, ops.Admin)
	assert.Contains(t, ops.Allowed, "run")
	assert.NotContains(t, ops.Allowed, "secret/view")
	assert.Contains(t, ops.Summary, "everything else is denied")
	assert.Equal(t, 3, ops.FolderCount)
	assert.Len(t, ops.Folders, 3)
	var artifact api.AreaView
	for _, area := range ops.Sections {
		if area.Key == "artifact" {
			artifact = area
		}
	}
	assert.Equal(t, "Artifacts", artifact.Title)
	assert.Equal(t, "Artifacts are the files runs leave.", artifact.About)
	assert.Equal(t, []string{"view", "edit"}, artifact.Rules)
	assert.True(t, artifact.Allowed)
	assert.Equal(t, "run", ops.Permissions[0].Key)
	assert.True(t, ops.Permissions[0].Allowed)

	viewer := meOf(t, f, asViewer)
	assert.NotContains(t, viewer.Allowed, "run")
	assert.False(t, viewer.Permissions[0].Allowed)

	limited := meOf(t, f, asLimited)
	assert.Equal(t, []string{"infra"}, limited.Folders)

	admin := meOf(t, f, asAdmin)
	assert.True(t, admin.Admin)
	assert.Equal(t, 100, api.MaxFolders)
	assert.Equal(t, "everything, as an admin", admin.Summary)
}

// An area marshals its rules only when it has any.
func TestAreaView(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(api.AreaView{Key: "secret", Title: "Secrets"})
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"secret","title":"Secrets","about":"","allowed":false}`, string(raw))
}
