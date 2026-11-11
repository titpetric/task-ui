//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A key's row is unique by name and by fingerprint; its links go with
// it.
func TestKeyInfoStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	keys := storage.NewKeyInfoStorage(newDB(t))
	row := &model.KeyInfo{Name: "deploy_rsa4096", Purpose: storage.KeySSH, Algo: "rsa", Bits: 4096, Fingerprint: "SHA256:a"}
	assert.NoError(t, keys.Create(ctx, row))
	assert.Equal(t, storage.ScopeInstance, row.Scope)
	assert.ErrorIs(t, keys.Create(ctx, &model.KeyInfo{Name: "deploy_rsa4096", Purpose: storage.KeySSH, Fingerprint: "SHA256:b"}), storage.ErrKeyExists)
	assert.ErrorIs(t, keys.Create(ctx, &model.KeyInfo{Name: "other_ed25519", Purpose: storage.KeyEncrypt, Fingerprint: "SHA256:a"}), storage.ErrKeyExists)
	assert.Error(t, keys.Create(ctx, &model.KeyInfo{Name: "x", Purpose: "both", Fingerprint: "SHA256:c"}))

	got, err := keys.Get(ctx, "deploy_rsa4096")
	assert.NoError(t, err)
	assert.Equal(t, int64(4096), got.Bits)
	byFP, err := keys.ByFingerprint(ctx, "SHA256:a")
	assert.NoError(t, err)
	assert.Equal(t, "deploy_rsa4096", byFP.Name)
	assert.NoError(t, keys.SetFingerprint(ctx, "deploy_rsa4096", "SHA256:z"))
	got, _ = keys.Get(ctx, "deploy_rsa4096")
	assert.Equal(t, "SHA256:z", got.Fingerprint)

	assert.NoError(t, keys.Link(ctx, "deploy_rsa4096", []string{"runner", "admin", "admin"}, []string{"claude", ""}))
	assert.Error(t, keys.Link(ctx, "deploy_rsa4096", []string{"root"}, nil))
	roles, users, err := keys.Links(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin", "runner"}, roles["deploy_rsa4096"])
	assert.Equal(t, []string{"claude"}, users["deploy_rsa4096"])

	list, err := keys.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.NoError(t, keys.Delete(ctx, "deploy_rsa4096"))
	_, err = keys.Get(ctx, "deploy_rsa4096")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	roles, _, _ = keys.Links(ctx)
	assert.Len(t, roles, 0)
}

// A rotation records the new fingerprint and the backup of the old
// material; clearing the backup keeps the fingerprint.
func TestKeyInfoStorage_SetBackup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	keys := storage.NewKeyInfoStorage(newDB(t))
	assert.NoError(t, keys.Create(ctx, &model.KeyInfo{Name: "task_ssh_ed25519", Purpose: storage.KeySSH, Algo: "ed25519", Fingerprint: "SHA256:old"}))
	assert.NoError(t, keys.SetBackup(ctx, "task_ssh_ed25519", "SHA256:new", "task_ssh_ed25519.bak", "SHA256:old"))
	got, err := keys.Get(ctx, "task_ssh_ed25519")
	assert.NoError(t, err)
	assert.Equal(t, "SHA256:new", got.Fingerprint)
	assert.Equal(t, "task_ssh_ed25519.bak", got.Backup)
	assert.Equal(t, "SHA256:old", got.BackupFingerprint)
	assert.NotNil(t, got.RotatedAt)
}

// ClearBackup forgets the backup and keeps the key's own fingerprint.
func TestKeyInfoStorage_ClearBackup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	keys := storage.NewKeyInfoStorage(newDB(t))
	assert.NoError(t, keys.Create(ctx, &model.KeyInfo{Name: "task_ssh_ed25519", Purpose: storage.KeySSH, Algo: "ed25519", Fingerprint: "SHA256:old"}))
	assert.NoError(t, keys.SetBackup(ctx, "task_ssh_ed25519", "SHA256:new", "task_ssh_ed25519.bak", "SHA256:old"))
	assert.NoError(t, keys.ClearBackup(ctx, "task_ssh_ed25519"))
	got, err := keys.Get(ctx, "task_ssh_ed25519")
	assert.NoError(t, err)
	assert.Equal(t, "", got.Backup)
	assert.Equal(t, "", got.BackupFingerprint)
	assert.Equal(t, "SHA256:new", got.Fingerprint)
}

// Every method returns the driver's error.
func TestKeyInfoStorage_Closed(t *testing.T) {
	t.Parallel()
	keys := storage.NewKeyInfoStorage(closedDB(t))
	ctx := context.Background()

	assert.ErrorContains(t, keys.Create(ctx, &model.KeyInfo{Name: "deploy", Purpose: storage.KeySSH}), errClosed)
	_, err := keys.Get(ctx, "deploy")
	assert.ErrorContains(t, err, errClosed)
	_, err = keys.ByFingerprint(ctx, "SHA256:x")
	assert.ErrorContains(t, err, errClosed)
	_, err = keys.List(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, keys.SetFingerprint(ctx, "deploy", "SHA256:x"), errClosed)
	assert.ErrorContains(t, keys.SetBackup(ctx, "deploy", "SHA256:x", "deploy.old", "SHA256:y"), errClosed)
	assert.ErrorContains(t, keys.ClearBackup(ctx, "deploy"), errClosed)
	assert.ErrorContains(t, keys.Delete(ctx, "deploy"), errClosed)
	assert.ErrorContains(t, keys.Link(ctx, "deploy", nil, nil), errClosed)
	_, _, err = keys.Links(ctx)
	assert.ErrorContains(t, err, errClosed)
}

// Delete and Link each run several statements in one transaction and
// stop at the one that failed, whichever it is: a table that is gone
// fails a delete, and a trigger fails an insert.
func TestKeyInfoStorage_Writes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, table := range []string{model.KeyRoleTable, model.KeyUserTable, model.KeyInfoTable} {
		keys := storage.NewKeyInfoStorage(brokenDB(t, table))
		assert.ErrorContains(t, keys.Delete(ctx, "deploy"), errNoTable, table)
	}
	for _, table := range []string{model.KeyRoleTable, model.KeyUserTable} {
		keys := storage.NewKeyInfoStorage(brokenDB(t, table))
		assert.ErrorContains(t, keys.Link(ctx, "deploy", nil, nil), errNoTable, table)
	}

	roles := newDB(t)
	refuse(t, roles, model.KeyRoleTable, "INSERT")
	assert.ErrorContains(t, storage.NewKeyInfoStorage(roles).Link(ctx, "deploy", []string{model.RoleAdmin}, nil), errRefused)

	users := newDB(t)
	refuse(t, users, model.KeyUserTable, "INSERT")
	assert.ErrorContains(t, storage.NewKeyInfoStorage(users).Link(ctx, "deploy", nil, []string{"alice"}), errRefused)
}

// Links reads two tables and returns the second read's error.
func TestKeyInfoStorage_LinksSecondRead(t *testing.T) {
	t.Parallel()
	_, _, err := storage.NewKeyInfoStorage(brokenDB(t, model.KeyUserTable)).Links(context.Background())
	assert.ErrorContains(t, err, errNoTable)
}
