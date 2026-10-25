package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A body signs with the secret as Forgejo signs it, and verifies only
// under the same secret.
func TestSign(t *testing.T) {
	t.Parallel()
	body := []byte(`{"ref":"refs/heads/main"}`)
	signature := forgejo.Sign("s3", body)
	assert.Len(t, signature, 64)
	// HMAC-SHA256("s3", body), as `openssl dgst -sha256 -hmac s3` prints it.
	assert.Equal(t, "aedc83c1fc378a8f6ea326bd6e8901c07a9c8a37f003a00c591939d577faaee9", signature)
	assert.True(t, forgejo.Verify("s3", body, signature))
	assert.False(t, forgejo.Verify("s4", body, signature))
	assert.False(t, forgejo.Verify("s3", []byte("other"), signature))
	assert.False(t, forgejo.Verify("s3", body, ""))
	assert.False(t, forgejo.Verify("", body, forgejo.Sign("", body)), "an empty secret verifies nothing")
}
