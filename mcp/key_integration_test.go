//go:build integration

package mcp_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/titpetric/task-ui/internal/assert"
)

// get_ssh_key returns an admin's agent an SSH key encrypted to its own
// age recipient, which opens to the key; whoami names the keys it may
// read. A runner without key/read gets a refusal, and nothing in clear.
func TestKeyTool(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["admin"], "/api/keys/generate", map[string]any{"name": "deploy", "policies": []string{"runner"}})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	identity, err := age.GenerateX25519Identity()
	assert.NoError(t, err)

	who, _ := f.tool(t, "admin", "whoami", nil)
	assert.Contains(t, who, "SSH keys you may read with get_ssh_key: deploy_ed25519")

	text, failed := f.tool(t, "admin", "get_ssh_key", map[string]any{"name": "deploy_ed25519", "recipient": identity.Recipient().String()})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Key deploy_ed25519 (ed25519, SHA256:")
	assert.NotContains(t, text, "OPENSSH PRIVATE KEY")
	opened, err := age.Decrypt(armor.NewReader(strings.NewReader(text[strings.Index(text, armor.Header):])), identity)
	assert.NoError(t, err)
	plain, err := io.ReadAll(opened)
	assert.NoError(t, err)
	assert.Contains(t, string(plain), "OPENSSH PRIVATE KEY")

	refused, failed := f.tool(t, "ops", "get_ssh_key", map[string]any{"name": "deploy_ed25519", "recipient": identity.Recipient().String()})
	assert.True(t, failed)
	assert.Contains(t, refused, "key/read")
	assert.NotContains(t, refused, "PRIVATE KEY")
	bad, failed := f.tool(t, "admin", "get_ssh_key", map[string]any{"name": "deploy_ed25519", "recipient": "nope"})
	assert.True(t, failed)
	assert.Contains(t, bad, "recipient")

	status, raw = f.send(t, http.MethodGet, f.tokens["admin"], "/api/keys", nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(raw), `"via":"mcp"`)
}
