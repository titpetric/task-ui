package git

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/sshagent"
)

// fakeKeys hand git a key in an agent, with key, or none.
type fakeKeys struct {
	t     *testing.T
	key   bool
	agent *sshagent.Agent
}

func (f *fakeKeys) GitAgent(context.Context) (*sshagent.Agent, string, error) {
	if !f.key {
		return nil, "", nil
	}
	_, private, _ := ed25519.GenerateKey(rand.Reader)
	block, err := ssh.MarshalPrivateKey(private, "")
	if err != nil {
		f.t.Fatal(err)
	}
	f.agent, err = sshagent.LoadKeys(f.t.TempDir(), sshagent.Named{Name: "task_ssh_ed25519", PEM: pem.EncodeToMemory(block)})
	return f.agent, "/keys/task_ssh_ed25519.pub", err
}

func (f *fakeKeys) KnownHosts() string { return "/keys/known_hosts" }

func TestGit_ValidURL(t *testing.T) {
	g := &Git{}
	for _, ok := range []string{"https://github.com/o/r.git", "ssh://git@host/r", "git@github.com:o/r.git", "file:///tmp/r.git", "git://host/r"} {
		if !g.ValidURL(ok) {
			t.Errorf("%q refused", ok)
		}
	}
	for _, bad := range []string{"", "-oops", "/tmp/r", "http//x", "git@github.com", "a b"} {
		if g.ValidURL(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestGit_IsRepo(t *testing.T) {
	g := &Git{}
	if g.IsRepo(t.TempDir()) {
		t.Fatal("empty folder is a repo")
	}
}
