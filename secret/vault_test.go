package secret_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
)

// newKeys returns an empty SSH folder.
func newKeys(t *testing.T) *sshkey.Store {
	t.Helper()
	return &sshkey.Store{Dir: t.TempDir()}
}

func TestNewVault(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)

	v := secret.NewVault(keys)
	assert.Equal(t, keys, v.Keys)
	assert.Nil(t, v.Recipients())
	_, err := v.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)
}

func TestVault_RoundTrip(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	_, err := keys.Generate("ed25519", "task-ui")
	assert.NoError(t, err)
	v := secret.NewVault(keys)

	assert.Equal(t, []string{"id_ed25519"}, v.Recipients())
	text, err := v.Encrypt("hunter2")
	assert.NoError(t, err)
	assert.True(t, secret.Encrypted(text))
	assert.NotContains(t, text, "hunter2")

	plain, err := v.Decrypt(text)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain)

	// Two encryptions of one value differ; both open.
	again, err := v.Encrypt("hunter2")
	assert.NoError(t, err)
	assert.NotEqual(t, text, again)
	plain, err = v.Decrypt(again)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain)

	// A value that was never encrypted comes back as it is.
	plain, err = v.Decrypt("plain")
	assert.NoError(t, err)
	assert.Equal(t, "plain", plain)
	assert.False(t, secret.Encrypted("plain"))

	_, err = v.Decrypt("-----BEGIN AGE ENCRYPTED FILE-----\ngarbage\n-----END AGE ENCRYPTED FILE-----\n")
	assert.Error(t, err)
}

// A rotation keeps the old pair as a backup: what it encrypted still
// opens, new values go to the new key only, and a backup never encrypts.
func TestVault_Rotation(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	_, err := keys.Generate("ed25519", "first")
	assert.NoError(t, err)
	v := secret.NewVault(keys)
	old, err := v.Encrypt("before")
	assert.NoError(t, err)

	_, err = keys.Generate("ed25519", "second")
	assert.NoError(t, err)
	assert.Len(t, keys.List(), 2)
	assert.Equal(t, []string{"id_ed25519"}, v.Recipients())

	plain, err := v.Decrypt(old)
	assert.NoError(t, err)
	assert.Equal(t, "before", plain)

	// Without the backup the old value is lost; a rekeyed one is not.
	fresh, err := v.Encrypt("after")
	assert.NoError(t, err)
	for _, key := range keys.List() {
		if key.Backup {
			assert.NoError(t, keys.Delete(key.Name))
		}
	}
	_, err = v.Decrypt(old)
	assert.Error(t, err)
	plain, err = v.Decrypt(fresh)
	assert.NoError(t, err)
	assert.Equal(t, "after", plain)

	// Excluding the only key leaves nothing to encrypt with.
	assert.Nil(t, v.Recipients("id_ed25519"))
	_, err = v.Encrypt("x", "id_ed25519")
	assert.ErrorIs(t, err, secret.ErrNoKey)
}

// ECDSA keys cannot encrypt; RSA keys can.
func TestVault_KeyTypes(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	_, err := keys.Generate("ecdsa", "curve")
	assert.NoError(t, err)
	v := secret.NewVault(keys)
	assert.Nil(t, v.Recipients())
	_, err = v.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)

	_, err = keys.Generate("rsa", "big")
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_rsa"}, v.Recipients())
	text, err := v.Encrypt("with rsa")
	assert.NoError(t, err)
	plain, err := v.Decrypt(text)
	assert.NoError(t, err)
	assert.Equal(t, "with rsa", plain)
}
