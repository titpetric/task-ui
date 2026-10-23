package keystore_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

// A key says its algorithm and size as its name ends with them, its
// workspace, and whether it encrypts.
func TestKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "rsa4096", keystore.Key{Algo: "rsa", Bits: 4096}.Kind())
	assert.Equal(t, "ed25519", keystore.Key{Algo: "ed25519"}.Kind())
	assert.Equal(t, "ecdsa384", keystore.Key{Algo: "ecdsa", Bits: 384}.Kind())
	assert.Equal(t, "w1", keystore.Key{Scope: keystore.ScopeWorkspace("w1")}.Workspace())
	assert.Equal(t, "", keystore.Key{Scope: "instance"}.Workspace())
	assert.True(t, keystore.Key{Purpose: keystore.PurposeEncrypt}.Encrypts())
	assert.False(t, keystore.Key{Purpose: keystore.PurposeSSH}.Encrypts())
}
