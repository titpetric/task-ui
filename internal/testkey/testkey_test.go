package testkey_test

import (
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testkey"
)

// The key parses as a private key of the size asked for, and the second
// call answers the same bytes: the point is that nothing generates one
// twice.
func TestRSA(t *testing.T) {
	t.Parallel()
	data := testkey.RSA(t)

	key, err := ssh.ParseRawPrivateKey(data)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	assert.Equal(t, string(data), string(testkey.RSA(t)), "the same key, not another")
	assert.NotEqual(t, string(data), string(testkey.Of(t, 2048)), "a size of its own")

	signer, err := ssh.ParsePrivateKey(data)
	assert.NoError(t, err)
	assert.Equal(t, "ssh-rsa", signer.PublicKey().Type())
}
