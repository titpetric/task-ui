package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/workspace"
)

// EnsureInstance makes the instance's encryption key once, private, in
// Data/instance/encrypt; RotateInstance keeps the previous one as a
// backup.
func TestKeyring_EnsureInstance(t *testing.T) {
	t.Parallel()
	k := &workspace.Keyring{Data: t.TempDir()}
	assert.Nil(t, k.InstanceKey())
	made, err := k.EnsureInstance("task-ui instance")
	assert.NoError(t, err)
	assert.True(t, made)
	key := k.InstanceKey()
	assert.NotNil(t, key)
	assert.Equal(t, "ssh-ed25519", key.Type)
	assert.Equal(t, filepath.Join(k.Data, "instance", "encrypt"), k.Instance().Dir)
	for _, dir := range []string{k.InstanceDir(), k.Instance().Dir} {
		assert.Equal(t, os.FileMode(0o700), mode(t, dir), dir)
	}
	for _, name := range []string{"id_ed25519", "id_ed25519.pub"} {
		assert.Equal(t, os.FileMode(0o600), mode(t, filepath.Join(k.Instance().Dir, name)), name)
	}

	made, err = k.EnsureInstance("again")
	assert.NoError(t, err)
	assert.False(t, made)
	assert.Equal(t, key.Fingerprint, k.InstanceKey().Fingerprint)

	rotated, err := k.RotateInstance("rotated")
	assert.NoError(t, err)
	assert.NotEqual(t, key.Fingerprint, rotated.Fingerprint)
	assert.Equal(t, rotated.Fingerprint, k.InstanceKey().Fingerprint)
	keys := k.Instance().List()
	assert.Len(t, keys, 2)
	assert.True(t, keys[1].Backup)
	assert.Equal(t, key.Fingerprint, keys[1].Fingerprint)
}
