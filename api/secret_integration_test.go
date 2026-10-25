//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

func TestSecret_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/secrets"},
		{http.MethodPost, "/secrets"},
		{http.MethodDelete, "/secrets/" + model.NewID()},
		{http.MethodGet, "/secrets/" + model.NewID() + "/reads"},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

func TestSecret_Set(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	res := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "API_TOKEN", Content: "hunter2"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))

	var view api.SecretView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "workspace:_", view.Scope, "no scope is the default workspace's shared one")
	assert.Equal(t, "API_TOKEN", view.Name)
	assert.True(t, model.ValidID(view.ID))
	assert.False(t, string(res.Raw) != "" && contains(res.Raw, "hunter2"))

	secrets, err := f.opts.Secrets.For(context.Background(), "infra")
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", secrets["API_TOKEN"])

	// Setting again replaces the content and keeps the row.
	again := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: " all ", Name: "API_TOKEN", Content: "changed"})
	assert.Equal(t, http.StatusOK, again.Status)
	assert.Equal(t, view.ID, again.str("id"))
	secrets, _ = f.opts.Secrets.For(context.Background(), "infra")
	assert.Equal(t, "changed", secrets["API_TOKEN"])

	scoped := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "DB_PASS", Content: "x"})
	assert.Equal(t, http.StatusOK, scoped.Status)
	assert.Equal(t, "infra", scoped.str("scope"))
}

func TestSecret_SetRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	folder := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "nowhere", Name: "X", Content: "x"})
	assert.Equal(t, http.StatusBadRequest, folder.Status)
	assert.Contains(t, folder.str("error"), "nowhere")

	name := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "bad name", Content: "x"})
	assert.Equal(t, http.StatusBadRequest, name.Status)
}

func TestSecret_ListAndDelete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	set := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "hunter2"})
	f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "TOKEN", Content: "other"})

	list := f.call(t, asAdmin, http.MethodGet, "/secrets", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("secrets"), 2)
	assert.False(t, contains(list.Raw, "hunter2"))
	assert.False(t, contains(list.Raw, "content"))

	deleted := f.call(t, asAdmin, http.MethodDelete, "/secrets/"+set.str("id"), nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/secrets", nil).list("secrets"), 1)

	// Deleting again succeeds and does nothing.
	again := f.call(t, asAdmin, http.MethodDelete, "/secrets/"+set.str("id"), nil)
	assert.Equal(t, http.StatusNoContent, again.Status)
}

// Without an encryption key the store keeps no secret, whatever SSH
// keys there are; a rekey writes every secret again for the current
// keys.
func TestSecret_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Keys.Rows.Delete(ctx, "task_encryption_ed25519"))

	none := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "x"})
	assert.Equal(t, http.StatusConflict, none.Status)
	assert.Contains(t, none.str("error"), "no encryption key")

	f.withKey(t)
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "x"}).Status, "an SSH key does not encrypt")
	_, err := f.opts.Keys.Generate(ctx, keystore.Spec{Name: "task_instance", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "x"}).Status)
	raw, err := f.opts.Secrets.Secrets.List(context.Background())
	assert.NoError(t, err)
	assert.True(t, secret.Encrypted(raw[0].Content))

	denied := f.call(t, asOps, http.MethodPost, "/secrets/rekey", map[string]any{})
	assert.Equal(t, http.StatusForbidden, denied.Status)
	rekeyed := f.call(t, asAdmin, http.MethodPost, "/secrets/rekey", map[string]any{})
	assert.Equal(t, http.StatusOK, rekeyed.Status, string(rekeyed.Raw))
	assert.Equal(t, float64(1), rekeyed.get("rekeyed"))
	assert.Len(t, rekeyed.list("keys"), 1)
}

// workspaceVaults finds the workspace of each secret's scope.
func (f *fixture) workspaceVaults() *secret.Workspaces {
	return &secret.Workspaces{Rows: f.opts.Workspaces}
}

// withWorkspaces keeps the fixture's secrets per workspace: the default
// one and acme, every secret on the instance's encryption key.
func (f *fixture) withWorkspaces(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme"}))
	f.opts.Secrets.Workspaces = f.workspaceVaults()
}

// A secret scoped to a workspace reaches its folders; all is the
// default workspace's; the API refuses a workspace that is not there. A
// rotation of the instance's encryption key and a rekey write every
// secret again; removing the workspace waits for its secrets to go.
func TestSecret_WorkspaceScope(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withWorkspaces(t)

	all := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "all", Name: "ROOT", Content: "r"})
	assert.Equal(t, http.StatusOK, all.Status, string(all.Raw))
	assert.Equal(t, "workspace:_", all.str("scope"))
	acme := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "workspace:acme", Name: "ACME", Content: "a"})
	assert.Equal(t, http.StatusOK, acme.Status, string(acme.Raw))
	assert.Equal(t, "workspace:acme", acme.str("scope"))

	unknown := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "workspace:nope", Name: "X", Content: "x"})
	assert.Equal(t, http.StatusBadRequest, unknown.Status)
	assert.Contains(t, unknown.str("error"), "unknown workspace nope")

	infra, err := f.opts.Secrets.For(context.Background(), "infra")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"ROOT": "r"}, infra)
	inAcme, err := f.opts.Secrets.For(context.Background(), "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"ACME": "a"}, inAcme)

	// Rotating the instance's key writes both secrets again; the backup
	// can go.
	rotated := f.call(t, asAdmin, http.MethodPost, "/keys/encrypt/rotate", nil)
	assert.Equal(t, http.StatusCreated, rotated.Status, string(rotated.Raw))
	assert.Equal(t, float64(2), rotated.get("rekeyed"))
	key := f.opts.Keys.EncryptKey(context.Background(), storage.ScopeInstance)
	assert.Equal(t, keystore.BackupName(key.Name), key.Backup)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/"+key.Name+"/backup", nil).Status)
	inAcme, err = f.opts.Secrets.For(context.Background(), "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, "a", inAcme["ACME"])

	rekeyed := f.call(t, asAdmin, http.MethodPost, "/secrets/rekey", map[string]any{})
	assert.Equal(t, http.StatusOK, rekeyed.Status, string(rekeyed.Raw))
	assert.Equal(t, float64(2), rekeyed.get("rekeyed"))

	// Removing acme waits for its secrets to go first, then takes the workspace.
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodDelete, "/workspaces/acme", nil).Status)
	held, err := f.opts.Secrets.Encrypted(context.Background(), "acme")
	assert.NoError(t, err)
	assert.Len(t, held, 1)
	assert.NoError(t, f.opts.Secrets.Delete(context.Background(), held[0].ID))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/acme", nil).Status)
	left, err := f.opts.Secrets.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, left, 1)
	assert.Equal(t, "ROOT", left[0].Name)
}

// A workspace needs no key of its own: the instance's encryption key
// stores its secrets as soon as it exists, and an encryption key for
// the workspace answers 400.
func TestSecret_NoKey(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme"}))
	f.opts.Secrets.Workspaces = f.workspaceVaults()

	stored := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "workspace:acme", Name: "TOKEN", Content: "x"})
	assert.Equal(t, http.StatusOK, stored.Status, string(stored.Raw))
	refused := f.call(t, asAdmin, http.MethodPost, "/workspaces/acme/keys/generate", api.WorkspaceKeyRequest{Encrypt: true})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
	assert.Equal(t, api.ErrWorkspaceEncrypt.Error(), refused.str("error"))
	inAcme, err := f.opts.Secrets.For(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"TOKEN": "x"}, inAcme)
}

// The list says how many runs read each secret and the last one; the
// log lists the reads, newest first, names only.
func TestSecret_Reads(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	token := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "API_TOKEN", Content: "hunter2"}).str("id")
	region := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "REGION", Content: "eu-central"}).str("id")
	for _, run := range []string{model.NewID(), model.NewID()} {
		assert.NoError(t, f.opts.SecretReads.Record(ctx, []model.SecretRead{
			{SecretID: token, Name: "API_TOKEN", RunID: run, Folder: "infra", Job: "deploy", UserName: asOps},
			{SecretID: region, Name: "REGION", RunID: run, Folder: "infra", Job: "deploy", UserName: asOps},
		}))
	}
	last := model.NewID()
	assert.NoError(t, f.opts.SecretReads.Record(ctx, []model.SecretRead{{SecretID: token, Name: "API_TOKEN", RunID: last, Folder: "infra", Dir: "edge", Job: "plan", UserName: asAdmin}}))

	list := f.call(t, asAdmin, http.MethodGet, "/secrets", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.False(t, contains(list.Raw, "hunter2") || contains(list.Raw, "eu-central"))
	var body struct{ Secrets []api.SecretView }
	assert.NoError(t, json.Unmarshal(list.Raw, &body))
	byName := map[string]api.SecretView{}
	for _, view := range body.Secrets {
		byName[view.Name] = view
	}
	assert.Equal(t, 3, byName["API_TOKEN"].Reads)
	assert.Equal(t, last, byName["API_TOKEN"].LastRead.RunID)
	assert.Equal(t, "edge", byName["API_TOKEN"].LastRead.Dir)
	assert.Equal(t, 2, byName["REGION"].Reads)

	reads := f.call(t, asAdmin, http.MethodGet, "/secrets/"+token+"/reads?limit=2", nil)
	assert.Equal(t, http.StatusOK, reads.Status, string(reads.Raw))
	assert.Len(t, reads.list("reads"), 2)
	assert.Equal(t, last, reads.list("reads")[0].(map[string]any)["run_id"])
	assert.False(t, contains(reads.Raw, "hunter2"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/secrets/"+model.NewID()+"/reads", nil).Status)
}
