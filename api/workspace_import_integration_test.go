//go:build integration

package api_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workspace"
)

const passphrase = "correct horse battery"

// acme makes the acme workspace on a fixture, its home a clone of the
// repository, the api project cloned inside it, an SSH key granted to ops and to a
// user the other instance does not have, a shared secret, a folder
// secret and an env file. It returns the workspace's ID.
func (f *fixture) acme(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	repo := "file://" + localRepo(t)
	res := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "acme", Branch: "main", Title: "Acme", Recursive: new(false)})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	id := res.str("workspace", "id")
	res = f.call(t, asAdmin, http.MethodPost, "/projects", api.ProjectCloneRequest{URL: repo, ID: "api", Workspace: "acme"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	res = f.call(t, asAdmin, http.MethodPost, "/workspaces/acme/keys/generate", api.WorkspaceKeyRequest{Type: "ed25519", Name: "deploy", Users: []string{asOps, "ghost"}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "TOKEN", Content: "shared value"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "acme/api", Name: "OWN", Content: "own value"}))
	assert.NoError(t, f.opts.Env.Managed.Set(ctx, &model.EnvFile{Folder: "acme/api", Path: ".env", Content: "MODE=prod\n"}))
	return id
}

// A workspace exported from one instance imports on another under a
// new name: the keys byte for byte, the secrets open with them, the
// env file, the grants, the home and the project come along under the
// new name.
func TestWorkspaceImport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	from, to := newFixture(t), newFixture(t)
	from.withKeyring(t)
	to.withKeyring(t)
	fromID := from.acme(t)
	bundle := from.export(t, "acme", passphrase)

	res := to.importBundle(t, asAdmin, bundle, passphrase, "acme2")
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "acme2", res.str("workspace", "folder"))
	assert.Equal(t, "Acme", res.str("workspace", "title"))
	assert.Equal(t, true, res.get("workspace", "shallow"), "Recursive off comes along")
	toID := res.str("workspace", "id")
	projects := res.list("projects")
	assert.Len(t, projects, 1)
	project := projects[0].(map[string]any)
	assert.Equal(t, "acme2/api", project["folder"])
	assert.Equal(t, true, project["cloned"], project["output"])
	assert.Equal(t, "acme2/home", res.str("home", "folder"))
	assert.Equal(t, true, res.get("home", "cloned"), res.str("home", "output"))
	assert.Equal(t, []any{"ghost"}, res.get("unknown_users"))
	assert.NoFileExists(t, filepath.Join(to.root, "acme2", "Makefile"), "the workspace is a plain directory")
	assert.FileExists(t, filepath.Join(to.root, "acme2", "home", "Makefile"))
	assert.FileExists(t, filepath.Join(to.root, "acme2", "src", "api", "Makefile"))
	imported, err := to.opts.Workspaces.GetByFolder(ctx, "acme2")
	assert.NoError(t, err)
	assert.Contains(t, imported.HomeURL, "file://", "the home repository comes along")
	assert.Equal(t, "main", imported.HomeBranch)

	want, _, _, err := from.opts.Keys.Bundle(ctx, keystore.ScopeWorkspace(fromID))
	assert.NoError(t, err)
	got, _, users, err := to.opts.Keys.Bundle(ctx, keystore.ScopeWorkspace(toID))
	assert.NoError(t, err)
	assert.Len(t, got, len(want))
	for name, data := range want {
		name = strings.Replace(name, "workspace_acme_", "workspace_acme2_", 1)
		assert.True(t, bytes.Equal(data, got[name]), name)
		info, err := os.Stat(filepath.Join(to.opts.Keys.Dir, name))
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), name)
	}

	secrets, err := to.opts.Secrets.For(ctx, "acme2/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"TOKEN": "shared value", "OWN": "own value"}, secrets)
	file, err := to.opts.Env.Managed.Get(ctx, model.EnvFile{Folder: "acme2/api", Path: ".env"})
	assert.NoError(t, err)
	assert.Equal(t, "MODE=prod\n", file.Content)
	written, err := os.ReadFile(filepath.Join(to.root, "acme2", "src", "api", ".env"))
	assert.NoError(t, err)
	assert.Contains(t, string(written), "MODE=prod")
	assert.Contains(t, string(written), `TOKEN=`)
	assert.Equal(t, []string{"ghost", asOps}, users["workspace_acme2_deploy_ed25519"])
	row, err := to.opts.Projects.GetByFolder(ctx, "acme2/api")
	assert.NoError(t, err)
	assert.Contains(t, row.SourceURL, "file://")

	// Over a folder held, the import answers 409. The bundle's own name
	// is free on this instance, but its keys are acme2's, and a key has
	// one holder: 409, before the import clones anything.
	held := to.importBundle(t, asAdmin, bundle, passphrase, "acme2")
	assert.Equal(t, http.StatusConflict, held.Status, string(held.Raw))
	same := to.importBundle(t, asAdmin, bundle, passphrase, "")
	assert.Equal(t, http.StatusConflict, same.Status, string(same.Raw))
	assert.Contains(t, same.str("error"), "is there already")
	assert.NoDirExists(t, filepath.Join(to.root, "acme"))

	// A removed workspace takes its env files and its keys with it;
	// importing it again brings them back. Its secrets go first, by
	// hand: a delete refuses while any is there.
	secretsHeld, err := to.opts.Secrets.Encrypted(ctx, "acme2")
	assert.NoError(t, err)
	assert.NotEmpty(t, secretsHeld)
	for _, secret := range secretsHeld {
		assert.NoError(t, to.opts.Secrets.Delete(ctx, secret.ID))
	}
	removed := to.call(t, asAdmin, http.MethodDelete, "/workspaces/acme2", nil)
	assert.Equal(t, http.StatusNoContent, removed.Status, string(removed.Raw))
	_, err = to.opts.Env.Managed.Get(ctx, model.EnvFile{Folder: "acme2/api", Path: ".env"})
	assert.ErrorIs(t, err, storage.ErrNotFound)
	again := to.importBundle(t, asAdmin, bundle, passphrase, "")
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))
	assert.Equal(t, "acme", again.str("workspace", "folder"))
	kept, err := to.opts.Env.Managed.Get(ctx, model.EnvFile{Folder: "acme/api", Path: ".env"})
	assert.NoError(t, err)
	assert.Equal(t, "MODE=prod\n", kept.Content)
	secrets, err = to.opts.Secrets.For(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, "own value", secrets["OWN"])
}

// nothing reports whether an import left anything of a folder behind:
// the checkout, the row, a project or a secret.
func nothing(t *testing.T, f *fixture, name string) {
	t.Helper()
	ctx := context.Background()
	assert.NoDirExists(t, filepath.Join(f.root, name))
	_, err := f.opts.Workspaces.GetByFolder(ctx, name)
	assert.Error(t, err, name+" has no row")
	projects, err := f.opts.Projects.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, projects, 1, "acme's project alone: the home is no project")
	secrets, err := f.opts.Secrets.Secrets.List(ctx)
	assert.NoError(t, err)
	for _, s := range secrets {
		assert.NotEqual(t, name, s.Scope)
	}
	keys, err := f.opts.Keys.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 3, "the instance's, _'s and acme's SSH key alone")
}

// A wrong passphrase and a hand made tar that leaves .ssh/ answer 400
// and leave nothing; a bundle whose SSH key a workspace here holds
// answers 409 and leaves nothing.
func TestWorkspaceImport_Refused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKeyring(t)
	f.acme(t)
	bundle := f.export(t, "acme", passphrase)

	wrong := f.importBundle(t, asAdmin, bundle, "wrong", "acme3")
	assert.Equal(t, http.StatusBadRequest, wrong.Status, string(wrong.Raw))
	assert.Contains(t, wrong.str("error"), "passphrase")
	nothing(t, f, "acme3")

	none := f.importBundle(t, asAdmin, bundle, "", "acme3")
	assert.Equal(t, http.StatusBadRequest, none.Status)

	evil := f.importBundle(t, asAdmin, handMade(t, map[string]string{"manifest.json": `{"format":3,"workspace":{"folder":"evil"}}`, "../escape": "x"}), passphrase, "acme3")
	assert.Equal(t, http.StatusBadRequest, evil.Status, string(evil.Raw))
	assert.Contains(t, evil.str("error"), "../escape")
	assert.NoFileExists(t, filepath.Join(filepath.Dir(f.root), "escape"))
	nothing(t, f, "acme3")

	held := f.importBundle(t, asAdmin, bundle, passphrase, "acme3")
	assert.Equal(t, http.StatusConflict, held.Status, string(held.Raw))
	nothing(t, f, "acme3")

	reserved := f.importBundle(t, asAdmin, bundle, passphrase, "internal")
	assert.Equal(t, http.StatusBadRequest, reserved.Status)

	big := f.importBundle(t, asAdmin, bytes.Repeat([]byte("x"), workspace.MaxBundle+2<<20), passphrase, "acme3")
	assert.Equal(t, http.StatusRequestEntityTooLarge, big.Status, string(big.Raw))
	nothing(t, f, "acme3")
}

// handMade encrypts a tar.gz of the entries to the passphrase.
func handMade(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	recipient, err := age.NewScryptRecipient(passphrase)
	assert.NoError(t, err)
	recipient.SetWorkFactor(10)
	var out bytes.Buffer
	encrypted, err := age.Encrypt(&out, recipient)
	assert.NoError(t, err)
	zipped := gzip.NewWriter(encrypted)
	archive := tar.NewWriter(zipped)
	for name, content := range entries {
		assert.NoError(t, archive.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}))
		_, err := archive.Write([]byte(content))
		assert.NoError(t, err)
	}
	assert.NoError(t, archive.Close())
	assert.NoError(t, zipped.Close())
	assert.NoError(t, encrypted.Close())
	return out.Bytes()
}

// The default workspace exports with its shared and folder secrets and
// its projects at the root, and imports as a named workspace: the
// shared scope is the new workspace's, each folder moves under it.
func TestWorkspaceImport_Default(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	from, to := newFixture(t), newFixture(t)
	from.withKeyring(t)
	to.withKeyring(t)
	from.acme(t)
	assert.NoError(t, from.opts.Secrets.Set(ctx, &model.Secret{Scope: "all", Name: "ROOT", Content: "root value"}))
	assert.NoError(t, from.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "INFRA", Content: "infra value"}))
	assert.NoError(t, from.opts.Secrets.Set(ctx, &model.Secret{Scope: "internal/claude-prompt", Name: "KEEP", Content: "instance"}))
	assert.NoError(t, from.opts.Env.Managed.Set(ctx, &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}))
	assert.NoError(t, from.opts.Projects.Create(ctx, &model.Project{Folder: "web", Title: "Web"}))
	bundle := from.export(t, "_", passphrase)

	read, err := workspace.ReadBundle(bytes.NewReader(bundle), passphrase)
	assert.NoError(t, err)
	assert.Len(t, read.Manifest.Secrets, 2, "neither acme's nor internal's")
	assert.Len(t, read.Manifest.Projects, 1)

	unnamed := to.importBundle(t, asAdmin, bundle, passphrase, "")
	assert.Equal(t, http.StatusBadRequest, unnamed.Status, string(unnamed.Raw))
	res := to.importBundle(t, asAdmin, bundle, passphrase, "legacy")
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "legacy", res.str("workspace", "title"))
	project := res.list("projects")[0].(map[string]any)
	assert.Equal(t, "legacy/web", project["folder"])
	assert.Equal(t, false, project["cloned"])
	assert.DirExists(t, layout.ProjectDir(to.root, "legacy", "web"))

	secrets, err := to.opts.Secrets.For(ctx, "legacy/infra")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"ROOT": "root value", "INFRA": "infra value"}, secrets)
	rows, err := to.opts.Secrets.Secrets.List(ctx)
	assert.NoError(t, err)
	scopes := []string{}
	for _, row := range rows {
		scopes = append(scopes, row.Scope)
	}
	assert.Equal(t, []string{"workspace:legacy", "legacy/infra"}, scopes)
	_, err = to.opts.Env.Managed.Get(ctx, model.EnvFile{Folder: "legacy/infra", Path: ".env"})
	assert.NoError(t, err)
}
