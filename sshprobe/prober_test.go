package sshprobe_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/sshagent"
	gossh "golang.org/x/crypto/ssh"
)

// agentKeys serve one generated key to a probe, in an agent under dir.
type agentKeys struct {
	t   *testing.T
	dir string
	mu  sync.Mutex
}

func (k *agentKeys) LoginAgent(context.Context) (*sshagent.Agent, error) {
	k.mu.Lock()
	if k.dir == "" {
		k.dir = k.t.TempDir()
	}
	dir := k.dir
	k.mu.Unlock()
	_, private, _ := ed25519.GenerateKey(rand.Reader)
	block, err := gossh.MarshalPrivateKey(private, "")
	if err != nil {
		return nil, err
	}
	return sshagent.LoadKeys(dir, sshagent.Named{Name: "task_ssh_ed25519", PEM: pem.EncodeToMemory(block)})
}
