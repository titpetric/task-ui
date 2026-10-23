//go:build integration

package keystore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// The keys of the older layout go into the folder once, named, with
// their policies, the git key set, and the old files left in place; what
// the old encryption keys encrypted opens with the new ones.
func TestLocal_Migrate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	data, dir := t.TempDir(), t.TempDir()
	l := &keystore.Local{Dir: dir, Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db), Tmp: t.TempDir()}

	instance := &sshkey.Store{Dir: filepath.Join(data, "instance", "encrypt")}
	_, err = instance.Generate("ed25519", "task-ui instance")
	assert.NoError(t, err)
	sealed, err := secret.NewVault(secret.DirKeys{Store: instance}).Encrypt("integration token")
	assert.NoError(t, err)
	wsEncrypt := &sshkey.Store{Dir: filepath.Join(data, "workspaces", "w1", "encrypt")}
	_, err = wsEncrypt.Generate("ed25519", "task-ui malta")
	assert.NoError(t, err)
	wsSecret, err := secret.NewVault(secret.DirKeys{Store: wsEncrypt}).Encrypt("workspace secret")
	assert.NoError(t, err)
	wsSSH := &sshkey.Store{Dir: filepath.Join(data, "workspaces", "w1", "ssh")}
	_, err = wsSSH.GenerateAs("ed25519", "build", "")
	assert.NoError(t, err)
	grants := storage.NewWorkspaceKeyGrantStorage(db)
	assert.NoError(t, grants.Set(ctx, "w1", "id_ed25519_build", []string{"runner"}, nil))

	ssh := &sshkey.Store{Dir: dir}
	own, err := ssh.Generate("ed25519", "task@task.cdn.si")
	assert.NoError(t, err)
	deploy, err := ssh.GenerateAs("ecdsa", "deploy", "root@pve5386")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "known_hosts"), []byte("git.example ssh-ed25519 AAAA\n"), 0o600))
	policies := storage.NewSSHKeyPolicyStorage(db)
	assert.NoError(t, policies.Set(ctx, "id_ecdsa_deploy", []string{"admin"}))
	assert.NoError(t, policies.SetUsers(ctx, "id_ecdsa_deploy", []string{"claude"}))

	legacy := keystore.Legacy{Data: data, Workspaces: []model.Workspace{{ID: "w1", Folder: "malta"}, {ID: "w0", Folder: model.Default}}, Policies: policies, Grants: grants}
	moved, err := l.Migrate(ctx, legacy)
	assert.NoError(t, err)
	names := map[string]string{}
	for _, m := range moved {
		names[m.Name] = m.Fingerprint
	}
	assert.Len(t, names, 5)
	assert.Equal(t, own.Fingerprint, names["task_ssh_ed25519"])
	assert.Equal(t, deploy.Fingerprint, names["deploy_ecdsa256"])
	assert.Contains(t, names, "task_encryption_ed25519")
	assert.Contains(t, names, "workspace_malta_encryption_ed25519")
	assert.Contains(t, names, "workspace_malta_build_ed25519")

	got, err := l.Get(ctx, "deploy_ecdsa256")
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin"}, got.Roles)
	assert.Equal(t, []string{"claude"}, got.Users)
	build, _ := l.Get(ctx, "workspace_malta_build_ed25519")
	assert.Equal(t, []string{"runner"}, build.Roles)
	assert.Equal(t, keystore.ScopeWorkspace("w1"), build.Scope)
	assert.Equal(t, "task_ssh_ed25519", l.Git(ctx).Name, "the type's own key pulls")
	assert.True(t, l.EncryptKey(ctx, storage.ScopeInstance).Encrypts())

	plain, err := secret.NewVault(l.Vault(storage.ScopeInstance)).Decrypt(sealed)
	assert.NoError(t, err)
	assert.Equal(t, "integration token", plain)
	plain, err = secret.NewVault(l.Vault(keystore.ScopeWorkspace("w1"))).Decrypt(wsSecret)
	assert.NoError(t, err)
	assert.Equal(t, "workspace secret", plain)

	assert.FileExists(t, filepath.Join(dir, "id_ed25519"), "the old files stay")
	assert.FileExists(t, filepath.Join(data, "instance", "encrypt", "id_ed25519"))
	assert.FileExists(t, filepath.Join(dir, "known_hosts"))

	again, err := l.Migrate(ctx, legacy)
	assert.NoError(t, err)
	assert.Len(t, again, 0, "once")
}

// A workspace's keys travel in a bundle by name with their links, and
// an import takes them under the workspace it makes; a bundle of keys
// the folder holds fails the check before anything clones.
func TestLocal_Bundle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	from := newStore(t)
	scope := keystore.ScopeWorkspace("w1")
	_, err := from.Generate(ctx, keystore.Spec{Name: "workspace_malta_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	assert.NoError(t, err)
	ssh, err := from.Generate(ctx, keystore.Spec{Name: "workspace_malta_ssh", Scope: scope})
	assert.NoError(t, err)
	assert.NoError(t, from.Link(ctx, ssh.Name, []string{"runner"}, []string{"ops"}))
	files, roles, users, err := from.Bundle(ctx, scope)
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "workspace_malta_ssh_ed25519.key")

	assert.ErrorIs(t, from.CheckBundle(ctx, files), keystore.ErrExists)
	to := newStore(t)
	assert.NoError(t, to.CheckBundle(ctx, files))
	taken, err := to.Restore(ctx, keystore.ScopeWorkspace("w9"), "malta", "malta2", files, roles, users)
	assert.NoError(t, err)
	assert.Len(t, taken, 1)
	got, err := to.Get(ctx, "workspace_malta2_ssh_ed25519")
	assert.NoError(t, err, "the workspace's name follows its folder")
	assert.Equal(t, []string{"ops"}, got.Users)
	assert.Equal(t, keystore.ScopeWorkspace("w9"), got.Scope)

	_, err = to.Restore(ctx, scope, "", "x", map[string][]byte{"other/x.key": []byte("x")}, nil, nil)
	assert.Error(t, err)
}
