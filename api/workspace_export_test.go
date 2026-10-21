package api_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/workspace"
)

// withKeyring turns on the secrets per workspace and makes the default
// workspace its row and its key, as the server does on start.
func (f *fixture) withKeyring(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	row, err := f.opts.Workspaces.GetByFolder(ctx, model.Default)
	assert.NoError(t, err)
	_, err = f.opts.WorkspaceKeys.Ensure(row.ID, "task-ui _")
	assert.NoError(t, err)
	f.opts.Secrets.Workspaces = &secret.Workspaces{Rows: f.opts.Workspaces, Keys: f.opts.WorkspaceKeys}
}

// export answers with a workspace's bundle.
func (f *fixture) export(t *testing.T, name, passphrase string) []byte {
	t.Helper()
	res := f.call(t, asAdmin, http.MethodPost, "/workspaces/"+name+"/export", api.WorkspaceExportRequest{Passphrase: passphrase})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, "application/octet-stream", res.Headers.Get("Content-Type"))
	assert.Contains(t, res.Headers.Get("Content-Disposition"), name+workspace.Extension)
	return res.Raw
}

// importBundle posts a bundle as the multipart form the import reads.
func (f *fixture) importBundle(t *testing.T, user string, bundle []byte, passphrase, folder string) response {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("bundle", "acme.taskui-ws")
	assert.NoError(t, err)
	_, err = part.Write(bundle)
	assert.NoError(t, err)
	assert.NoError(t, form.WriteField("passphrase", passphrase))
	assert.NoError(t, form.WriteField("folder", folder))
	assert.NoError(t, form.Close())
	req, err := http.NewRequest(http.MethodPost, f.server.URL+"/workspaces/import", &body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("X-User", user)
	return f.do(t, req)
}

func TestWorkspaceExport_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKeyring(t)
	for _, user := range []string{asOps, asViewer, asLimited} {
		res := f.call(t, user, http.MethodPost, "/workspaces/_/export", api.WorkspaceExportRequest{Passphrase: "p"})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
		res = f.importBundle(t, user, []byte("x"), "p", "acme")
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}
	res := f.call(t, asAdmin, http.MethodPost, "/workspaces/_/export", api.WorkspaceExportRequest{})
	assert.Equal(t, http.StatusBadRequest, res.Status, "a passphrase is required")
	res = f.call(t, asAdmin, http.MethodPost, "/workspaces/nope/export", api.WorkspaceExportRequest{Passphrase: "p"})
	assert.Equal(t, http.StatusNotFound, res.Status)
}

// A workspace named import keeps its own export route.
func TestWorkspaceExport_NamedImport(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKeyring(t)
	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: "import"}))
	assert.NotEmpty(t, f.export(t, "import", "p"))
}
