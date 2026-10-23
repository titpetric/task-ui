package secret_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/secret"
)

// A chain encrypts with its first set's key and decrypts with every
// set's: a secret of a workspace's old key opens with the legacy chain.
func TestChain(t *testing.T) {
	t.Parallel()
	instance, workspace := newKeys(t), newKeys(t)
	_, err := instance.Generate("ed25519", "instance")
	assert.NoError(t, err)
	_, err = workspace.Generate("ed25519", "workspace")
	assert.NoError(t, err)
	old, err := secret.NewVault(secret.DirKeys{Store: workspace}).Encrypt("hunter2")
	assert.NoError(t, err)

	chain := secret.Chain{secret.DirKeys{Store: instance}, secret.DirKeys{Store: workspace}}
	name, _, ok := chain.Recipient()
	assert.True(t, ok)
	assert.Equal(t, "id_ed25519", name)
	assert.Len(t, chain.Identities(), 2)
	plain, err := secret.NewVault(chain).Decrypt(old)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain)

	_, _, ok = secret.Chain{}.Recipient()
	assert.False(t, ok)
}
