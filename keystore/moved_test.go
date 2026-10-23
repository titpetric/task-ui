package keystore_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

// A moved key names its old file, its name and its fingerprint, never
// its material.
func TestMoved(t *testing.T) {
	t.Parallel()
	moved := keystore.Moved{From: "/keys/id_rsa_deploy", Name: "deploy_rsa4096", Fingerprint: "SHA256:x"}
	assert.Equal(t, "deploy_rsa4096", moved.Name)
}
