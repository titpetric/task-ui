package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// The tree lists a directory's latest files with the run or the user
// that wrote each, directories first; a runner limited to a folder sees
// that folder's files and the ones users wrote.
func TestArtifactTree(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.version(t, "infra", "reports/r.md", "# One\n")
	f.version(t, "private", "secret.txt", "hidden")
	assert.Equal(t, http.StatusCreated, f.put(t, asOps, "todo.txt", "write me").Status)

	var root api.ArtifactTreeView
	res := f.call(t, asAdmin, http.MethodGet, "/artifacts/tree", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.NoError(t, json.Unmarshal(res.Raw, &root))
	assert.Len(t, root.Entries, 3)
	assert.Equal(t, "reports", root.Entries[0].Name)
	assert.True(t, root.Entries[0].Dir)

	res = f.call(t, asLimited, http.MethodGet, "/artifacts/tree", nil)
	var limited api.ArtifactTreeView
	assert.NoError(t, json.Unmarshal(res.Raw, &limited))
	names := []string{}
	for _, entry := range limited.Entries {
		names = append(names, entry.Name)
	}
	assert.Equal(t, []string{"reports", "todo.txt"}, names)
	assert.Equal(t, "ops", limited.Entries[1].UserName)
	assert.Equal(t, int64(8), limited.Entries[1].ByteSize)

	var reports api.ArtifactTreeView
	res = f.call(t, asLimited, http.MethodGet, "/artifacts/tree/reports", nil)
	assert.NoError(t, json.Unmarshal(res.Raw, &reports))
	assert.Equal(t, "reports", reports.Path)
	assert.Len(t, reports.Entries, 1)
	entry := api.ArtifactEntry{Name: "r.md", Path: "reports/r.md", Folder: "infra", Job: "build"}
	assert.Equal(t, entry.Path, reports.Entries[0].Path)
	assert.Equal(t, entry.Folder, reports.Entries[0].Folder)
	assert.Equal(t, entry.Job, reports.Entries[0].Job)

	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/artifacts/tree/nothing", nil).Status)
}
