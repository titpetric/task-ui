//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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
