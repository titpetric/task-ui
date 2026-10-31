//go:build integration

package keystore_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testkey"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/storage"
)

// Generate names a key by its algorithm and size, writes its two files
// 0600 in a 0700 folder, and refuses a name taken, under any algorithm.
func TestLocal_Generate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	for _, c := range []struct {
		algo string
		bits int
		want string
	}{
		{"", 0, "a_ed25519"}, {"ecdsa", 256, "b_ecdsa256"}, {"ecdsa", 384, "c_ecdsa384"}, {"ecdsa", 521, "d_ecdsa521"},
		{"rsa", 2048, "e_rsa2048"}, {"rsa", 3072, "f_rsa3072"},
	} {
		key, err := l.Generate(ctx, keystore.Spec{Name: strings.SplitN(c.want, "_", 2)[0], Algo: c.algo, Bits: c.bits, Comment: "test"})
		assert.NoError(t, err)
		assert.Equal(t, c.want, key.Name)
		assert.Equal(t, keystore.PurposeSSH, key.Purpose)
		assert.Equal(t, storage.ScopeInstance, key.Scope)
		assert.True(t, strings.HasSuffix(key.PublicKey, " test"))
		for _, ext := range []string{".key", ".pub"} {
			info, err := os.Stat(filepath.Join(l.Dir, key.Name+ext))
			assert.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		}
	}
	info, err := os.Stat(l.Dir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())

	_, err = l.Import(ctx, keystore.Spec{Name: "a"}, testkey.Of(t, 2048))
	assert.ErrorIs(t, err, keystore.ErrExists, "a name is unique across algorithms")
	_, err = l.Generate(ctx, keystore.Spec{Name: "x", Algo: "ecdsa", Bits: 128})
	assert.Error(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "x", Algo: "dsa"})
	assert.Error(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "x", Algo: "ecdsa", Purpose: keystore.PurposeEncrypt})
	assert.ErrorIs(t, err, keystore.ErrEncrypt)
	_, err = l.Generate(ctx, keystore.Spec{Name: "x", Purpose: "both"})
	assert.ErrorIs(t, err, keystore.ErrPurpose)

	_, err = l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "second", Purpose: keystore.PurposeEncrypt})
	assert.ErrorIs(t, err, keystore.ErrExists, "one encryption key a scope")
	assert.Equal(t, "task_encryption_ed25519", l.EncryptKey(ctx, storage.ScopeInstance).Name)
}
