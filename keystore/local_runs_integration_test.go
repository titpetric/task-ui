//go:build integration

package keystore_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

// Links decide which runs get an SSH key: an instance key goes to the
// roles and users it links, a workspace's to every run of the workspace
// without links, to the linked ones with. An encryption key goes to no
// run and links nothing.
func TestLocal_ForRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	mustGen := func(spec keystore.Spec) *keystore.Key {
		key, err := l.Generate(ctx, spec)
		assert.NoError(t, err)
		return key
	}
	deploy := mustGen(keystore.Spec{Name: "deploy"})
	assert.NoError(t, l.Link(ctx, deploy.Name, []string{"admin"}, []string{"claude"}))
	mustGen(keystore.Spec{Name: "task_ssh"})
	ws := mustGen(keystore.Spec{Name: "workspace_malta_ssh", Scope: keystore.ScopeWorkspace("w1")})
	limited := mustGen(keystore.Spec{Name: "workspace_malta_build", Scope: keystore.ScopeWorkspace("w1")})
	assert.NoError(t, l.Link(ctx, limited.Name, nil, []string{"ops"}))
	enc := mustGen(keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.ErrorIs(t, l.Link(ctx, enc.Name, []string{"admin"}, nil), keystore.ErrNotSSH)

	names, err := l.ForRun(ctx, "admin", "black", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{deploy.Name}, names)
	names, _ = l.ForRun(ctx, "runner", "claude", "w1")
	assert.Equal(t, []string{deploy.Name, ws.Name}, names)
	names, _ = l.ForRun(ctx, "runner", "ops", "w1")
	assert.Equal(t, []string{limited.Name, ws.Name}, names)
	names, _ = l.ForRun(ctx, "viewer", "guest", "w2")
	assert.Len(t, names, 0)

	logins, err := l.Logins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{deploy.Name, "task_ssh_ed25519"}, logins)
	probe, err := l.LoginAgent(ctx)
	assert.NoError(t, err)
	assert.Equal(t, logins, probe.Keys, "a probe signs in with every instance SSH key")
	assert.NoError(t, probe.Close())

	_, err = l.Materialize(ctx, []string{enc.Name})
	assert.ErrorIs(t, err, keystore.ErrNotSSH, "an encryption key serves no run")
}

// A run's keys are in an agent, from memory: its directory holds the
// socket alone, and closing it leaves nothing.
func TestLocal_Materialize(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	key, err := l.Generate(ctx, keystore.Spec{Name: "deploy"})
	assert.NoError(t, err)
	a, err := l.Materialize(ctx, []string{key.Name})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(a.Socket, l.Tmp))
	entries, _ := os.ReadDir(filepath.Dir(a.Socket))
	assert.Len(t, entries, 1)
	assert.Equal(t, "agent.sock", entries[0].Name())
	conn, err := net.Dial("unix", a.Socket)
	assert.NoError(t, err)
	listed, err := agent.NewClient(conn).List()
	assert.NoError(t, err)
	assert.Len(t, listed, 1)
	pub, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(key.PublicKey))
	assert.Equal(t, ssh.FingerprintSHA256(pub), ssh.FingerprintSHA256(listed[0]))
	conn.Close()
	assert.NoError(t, a.Close())
	left, _ := os.ReadDir(l.Tmp)
	assert.Len(t, left, 0, "nothing on disk after the run")

	none, err := l.Materialize(ctx, nil)
	assert.NoError(t, err)
	assert.Nil(t, none)
	_, err = l.Materialize(ctx, []string{"missing"})
	assert.Error(t, err)
}
