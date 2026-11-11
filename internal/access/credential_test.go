package access_test

import (
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestCredential(t *testing.T) {
	t.Parallel()

	cred := access.Credential{Name: "alice", Hash: "$2a$10$x"}
	assert.Equal(t, "alice:$2a$10$x", cred.Line())
}
