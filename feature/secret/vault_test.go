package secret_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testkey"
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

	v := secret.NewVault(secret.DirKeys{Store: keys})
	assert.Equal(t, secret.Keys(secret.DirKeys{Store: keys}), v.Keys)
	assert.Nil(t, v.Recipients())
	_, err := v.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)
}

func TestVault_RoundTrip(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	_, err := keys.Generate("ed25519", "task-ui")
	assert.NoError(t, err)
	v := secret.NewVault(secret.DirKeys{Store: keys})

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

	// A plain value comes back unchanged.
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
	v := secret.NewVault(secret.DirKeys{Store: keys})
	old, err := v.Encrypt("before")
	assert.NoError(t, err)

	_, err = keys.Generate("ed25519", "second")
	assert.NoError(t, err)
	assert.Len(t, keys.List(), 2)
	assert.Equal(t, []string{"id_ed25519"}, v.Recipients())

	plain, err := v.Decrypt(old)
	assert.NoError(t, err)
	assert.Equal(t, "before", plain)

	// Without the backup the old value stays unreadable; a rekeyed one decrypts.
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
	v := secret.NewVault(secret.DirKeys{Store: keys})
	assert.Nil(t, v.Recipients())
	_, err = v.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)

	_, err = keys.Upload(testkey.RSA(t), "big")
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_rsa"}, v.Recipients())
	text, err := v.Encrypt("with rsa")
	assert.NoError(t, err)
	plain, err := v.Decrypt(text)
	assert.NoError(t, err)
	assert.Equal(t, "with rsa", plain)
}

// The secrets have one instance key: a type's own, Ed25519 before RSA. A
// key made for someone or for a use never encrypts them, and still opens
// what it encrypted before.
func TestVault_InstanceKey(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	deploy, err := keys.UploadAs(testkey.RSA(t), "deploy", "deploy")
	assert.NoError(t, err)
	v := secret.NewVault(secret.DirKeys{Store: keys})
	assert.Nil(t, v.Recipients())
	_, err = v.Encrypt("x")
	assert.ErrorIs(t, err, secret.ErrNoKey)

	_, err = keys.Upload(testkey.RSA(t), "own")
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_rsa"}, v.Recipients())
	_, err = keys.Generate("ed25519", "instance")
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_ed25519"}, v.Recipients())
	assert.Nil(t, v.Recipients("id_ed25519"), "an excluded key leaves the vault none")

	text, err := v.Encrypt("one key")
	assert.NoError(t, err)
	plain, err := v.Decrypt(text)
	assert.NoError(t, err)
	assert.Equal(t, "one key", plain)
	assert.Equal(t, []string{"ssh-ed25519"}, stanzas(t, text))
	assert.Equal(t, "id_rsa_deploy", deploy.Name)
}

// stanzas returns the kinds of recipient an armored age file opens for,
// from its header: one stanza a key that opens it.
// for, from its header: one stanza a key that opens it.
func stanzas(t *testing.T, text string) []string {
	t.Helper()
	var body strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		if !strings.HasPrefix(line, "-----") {
			body.WriteString(line)
		}
	}
	raw, err := base64.StdEncoding.DecodeString(body.String())
	assert.NoError(t, err)
	header, _, _ := strings.Cut(string(raw), "\n---")
	var kinds []string
	for _, line := range strings.Split(header, "\n") {
		if fields := strings.Fields(line); len(fields) > 1 && fields[0] == "->" {
			kinds = append(kinds, fields[1])
		}
	}
	return kinds
}
