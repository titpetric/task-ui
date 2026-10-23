package secret_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/secret"
)

// Keys in memory decrypt what their pair encrypted and encrypt nothing.
func TestPEMKeys(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	_, err := keys.Generate("ed25519", "bundle")
	assert.NoError(t, err)
	text, err := secret.NewVault(secret.DirKeys{Store: keys}).Encrypt("hunter2")
	assert.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(keys.Dir, "id_ed25519"))
	assert.NoError(t, err)

	keysInMemory := secret.PEMKeys{data}
	_, _, ok := keysInMemory.Recipient()
	assert.False(t, ok, "keys in memory never encrypt")
	assert.Len(t, keysInMemory.Identities(), 1)
	vault := secret.NewVault(keysInMemory)
	plain, err := vault.Decrypt(text)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain)
	_, err = vault.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)
}
