package keystore_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

// An export bundles a workspace's SSH keys alone, the private and the
// public half: no instance key, no encryption key, no other
// workspace's key. An import takes them under the folder it makes, with
// their links, and refuses a key the folder holds.
func TestLocal_BundleSSHOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newLocal(t)
	scope := keystore.ScopeWorkspace("w1")
	_, err := l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "task_ssh"})
	assert.NoError(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "workspace_acme_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	assert.NoError(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "workspace_beta_ssh", Scope: keystore.ScopeWorkspace("w2")})
	assert.NoError(t, err)
	key, err := l.Generate(ctx, keystore.Spec{Name: "workspace_acme_ssh", Scope: scope})
	assert.NoError(t, err)
	assert.NoError(t, l.Link(ctx, key.Name, []string{"runner"}, []string{"ops"}))

	files, roles, users, err := l.Bundle(ctx, scope)
	assert.NoError(t, err)
	names := []string{}
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	assert.Equal(t, []string{"workspace_acme_ssh_ed25519.key", "workspace_acme_ssh_ed25519.pub"}, names)
	assert.Equal(t, key.PublicKey, strings.TrimSpace(string(files["workspace_acme_ssh_ed25519.pub"])))
	assert.ErrorIs(t, l.CheckBundle(ctx, files), keystore.ErrExists)

	other := newLocal(t)
	assert.NoError(t, other.CheckBundle(ctx, files))
	taken, err := other.Restore(ctx, keystore.ScopeWorkspace("w9"), "acme", "gamma", files, roles, users)
	assert.NoError(t, err)
	assert.Equal(t, []string{"workspace_gamma_ssh_ed25519"}, taken)
	got, err := other.Get(ctx, "workspace_gamma_ssh_ed25519")
	assert.NoError(t, err)
	assert.Equal(t, key.Fingerprint, got.Fingerprint)
	assert.Equal(t, []string{"runner"}, got.Roles)
	assert.Equal(t, []string{"ops"}, got.Users)
	assert.Equal(t, keystore.ScopeWorkspace("w9"), got.Scope)

	_, err = other.Restore(ctx, keystore.ScopeWorkspace("w10"), "acme", "gamma", files, nil, nil)
	assert.ErrorIs(t, err, keystore.ErrExists, "a name the folder holds refuses the key")
	_, err = other.Restore(ctx, scope, "acme", "acme", map[string][]byte{"../escape.key": files["workspace_acme_ssh_ed25519.key"]}, nil, nil)
	assert.Error(t, err)
}
