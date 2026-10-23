//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"filippo.io/age"
	"filippo.io/age/armor"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// A private key goes only to an MCP agent user with key/read, for an
// SSH key linked to them or their role, any SSH key for an admin, and
// only encrypted to the age recipient the request names: the identity
// opens the ciphertext to the key. Each read goes to the log; no answer
// holds the key in clear, and an encryption key never leaves.
func TestKey_Read(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	deploy, err := f.opts.Keys.Generate(ctx, keystore.Spec{Name: "deploy"})
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Keys.Link(ctx, deploy.Name, []string{model.RoleRunner}, nil))
	other, err := f.opts.Keys.Generate(ctx, keystore.Spec{Name: "other"})
	assert.NoError(t, err)
	identity, err := age.GenerateX25519Identity()
	assert.NoError(t, err)
	recipient := url.QueryEscape(identity.Recipient().String())
	path := func(name string) string { return "/keys/" + name + "/private?recipient=" + recipient }
	private, err := f.opts.Keys.Private(ctx, deploy.Name)
	assert.NoError(t, err)
	clear := strings.TrimSpace(string(private))
	inClear := func(res response) {
		t.Helper()
		assert.NotContains(t, string(res.Raw), "PRIVATE KEY")
		assert.NotContains(t, string(res.Raw), clear[40:80])
	}

	// A runner by default, and a viewer, get 403: no key/read.
	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodGet, path(deploy.Name), nil)
		assert.Equal(t, http.StatusForbidden, res.Status, user)
		inClear(res)
	}
	// With key/read, still not without MCP agent user.
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asOps, "key/read", "allow"))
	ops := f.users[asOps]
	ops.Mcp = false
	assert.NoError(t, f.opts.Users.Update(ctx, ops))
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodGet, path(deploy.Name), nil).Status)
	ops.Mcp = true
	assert.NoError(t, f.opts.Users.Update(ctx, ops))

	res := f.call(t, asOps, http.MethodGet, path(deploy.Name), nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	inClear(res)
	assert.Equal(t, deploy.Fingerprint, res.str("fingerprint"))
	assert.Contains(t, res.str("note"), "age -d -i")
	var view api.KeyReadView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, deploy.Name, view.Name)
	opened, err := age.Decrypt(armor.NewReader(strings.NewReader(res.str("ciphertext"))), identity)
	assert.NoError(t, err)
	got, err := io.ReadAll(opened)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(private, got), "the identity opens it to the key")

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodGet, path(other.Name), nil).Status, "linked neither to ops nor to the runner role")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodGet, path("task_encryption_ed25519"), nil).Status, "an encryption key never leaves")
	assert.Equal(t, http.StatusForbidden, f.call(t, asAdmin, http.MethodGet, path("task_encryption_ed25519"), nil).Status, "not to an admin either")
	for _, bad := range []string{"", "age1nope", "ssh-ed25519 AAAA"} {
		res := f.call(t, asOps, http.MethodGet, "/keys/"+deploy.Name+"/private?recipient="+url.QueryEscape(bad), nil)
		assert.Equal(t, http.StatusBadRequest, res.Status, bad)
	}
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, path("missing"), nil).Status)

	admin := f.call(t, asAdmin, http.MethodGet, path(other.Name), nil)
	assert.Equal(t, http.StatusOK, admin.Status, "an admin reads any SSH key")

	reads, err := f.opts.KeyReads.All(ctx)
	assert.NoError(t, err)
	assert.Len(t, reads, 2)
	assert.Equal(t, "ops", reads[0].UserName)
	assert.Equal(t, "api", reads[0].Via)
	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Equal(t, float64(1), list.get("reads", deploy.Name))
	assert.Equal(t, "admin", list.str("last_read", other.Name, "user_name"))

	me := f.call(t, asOps, http.MethodGet, "/me", nil)
	assert.Equal(t, []any{deploy.Name}, me.get("keys"))
}
