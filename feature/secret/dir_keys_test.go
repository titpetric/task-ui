package secret_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/secret"
)

// A folder of id_* files encrypts with the type's own key, ed25519
// first, and decrypts with every private key in it.
func TestDirKeys(t *testing.T) {
	t.Parallel()
	keys := newKeys(t)
	d := secret.DirKeys{Store: keys}
	_, _, ok := d.Recipient()
	assert.False(t, ok)
	_, err := keys.GenerateAs("ed25519", "deploy", "")
	assert.NoError(t, err)
	_, _, ok = d.Recipient()
	assert.False(t, ok, "a key made for a use encrypts nothing")
	_, err = keys.Generate("ed25519", "own")
	assert.NoError(t, err)
	name, public, ok := d.Recipient()
	assert.True(t, ok)
	assert.Equal(t, "id_ed25519", name)
	assert.Contains(t, public, "ssh-ed25519")
	assert.Len(t, d.Identities(), 2)
}
