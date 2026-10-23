package keystore_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

func TestStem(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, algo string
		bits       int
		want       string
	}{
		{"deploy", "rsa", 4096, "deploy_rsa4096"},
		{"deploy_rsa4096", "rsa", 4096, "deploy_rsa4096"},
		{"task_ssh", "ed25519", 0, "task_ssh_ed25519"},
		{"edge", "ecdsa", 384, "edge_ecdsa384"},
	} {
		got, err := keystore.Stem(c.name, c.algo, c.bits)
		assert.NoError(t, err)
		assert.Equal(t, c.want, got)
	}
	for _, bad := range []string{"", "Deploy", "a b", "-x", "../x"} {
		_, err := keystore.Stem(bad, "ed25519", 0)
		assert.ErrorIs(t, err, keystore.ErrName, bad)
	}
}
