//go:build integration

package sshagent_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh/agent"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshagent"
	"github.com/titpetric/task-ui/sshkey"
)

// A directory as a checkout's .ssh/ holds it: a key pair, known_hosts and
// a file that is no key. The agent serves only the private key.
func TestLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := (&sshkey.Store{Dir: dir}).Generate("ed25519", "deploy")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "known_hosts"), []byte("example.com ssh-ed25519 AAAA\n"), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("not a key\n"), 0o600))

	a, err := sshagent.Load(dir, t.TempDir())
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, []string{filepath.Join(dir, "id_ed25519")}, a.Keys)

	conn, err := net.Dial("unix", a.Socket)
	assert.NoError(t, err)
	keys, err := agent.NewClient(conn).List()
	conn.Close()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "id_ed25519", keys[0].Comment)

	assert.NoError(t, a.Close())
	_, err = os.Stat(a.Socket)
	assert.True(t, os.IsNotExist(err))
}

// LoadFiles serves the keys of the files given, each once, from any
// directories, and skips what is not a key.
func TestLoadFiles(t *testing.T) {
	t.Parallel()
	one, two := t.TempDir(), t.TempDir()
	_, err := (&sshkey.Store{Dir: one}).Generate("ed25519", "project")
	assert.NoError(t, err)
	_, err = (&sshkey.Store{Dir: two}).Generate("rsa", "instance")
	assert.NoError(t, err)
	first, second := filepath.Join(one, "id_ed25519"), filepath.Join(two, "id_rsa")

	a, err := sshagent.LoadFiles(t.TempDir(), first, second, first, second+".pub", filepath.Join(two, "missing"))
	assert.NoError(t, err)
	defer a.Close()
	assert.Equal(t, []string{first, second}, a.Keys)

	conn, err := net.Dial("unix", a.Socket)
	assert.NoError(t, err)
	keys, err := agent.NewClient(conn).List()
	conn.Close()
	assert.NoError(t, err)
	assert.Len(t, keys, 2)

	none, err := sshagent.LoadFiles(t.TempDir())
	assert.NoError(t, err)
	assert.Nil(t, none)
}

// LoadKeys serves keys from memory under their names, and Clean
// removes the directories agents left.
func TestLoadKeys(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := (&sshkey.Store{Dir: dir}).Generate("ed25519", "deploy")
	assert.NoError(t, err)
	pem, err := os.ReadFile(filepath.Join(dir, "id_ed25519"))
	assert.NoError(t, err)
	tmp := t.TempDir()
	a, err := sshagent.LoadKeys(tmp, sshagent.Named{Name: "deploy_ed25519", PEM: pem}, sshagent.Named{Name: "deploy_ed25519", PEM: pem})
	assert.NoError(t, err)
	assert.Equal(t, []string{"deploy_ed25519"}, a.Keys)
	conn, err := net.Dial("unix", a.Socket)
	assert.NoError(t, err)
	keys, err := agent.NewClient(conn).List()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	conn.Close()
	entries, _ := os.ReadDir(filepath.Dir(a.Socket))
	assert.Len(t, entries, 1, "the socket alone, no key file")

	none, err := sshagent.LoadKeys(tmp)
	assert.NoError(t, err)
	assert.Nil(t, none)
	_, err = sshagent.LoadKeys(tmp, sshagent.Named{Name: "bad", PEM: []byte("x")})
	assert.Error(t, err)

	assert.Equal(t, 1, sshagent.Clean(tmp), "the agent's directory, as one left behind")
	assert.NoError(t, a.Close())
}
