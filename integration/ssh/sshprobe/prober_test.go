package sshprobe_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"sync"
	"testing"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/integration/ssh/sshagent"
	"github.com/titpetric/task-ui/integration/ssh/sshprobe"
	"github.com/titpetric/task-ui/internal/assert"
)

// agentKeys serve one generated key to a probe, in an agent under dir.
type agentKeys struct {
	t   *testing.T
	dir string
	mu  sync.Mutex
}

// A prober with no host probes nothing.
func TestProber(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Users: []string{"debian"}}
	assert.Empty(t, p.Logins(context.Background(), nil, time.Millisecond))
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
