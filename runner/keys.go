package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshagent"
)

// agent loads the private keys of the checkout's .ssh/ folder into an ssh
// agent for the run, and says so in the terminal. No checkout, no folder
// or no key in it is no agent and no error; the caller closes the agent.
func (m *Manager) agent(session *Session) (*sshagent.Agent, error) {
	dir := m.checkout(session)
	if dir == "" {
		return nil, nil
	}
	keys := filepath.Join(dir, ".ssh")
	if info, err := os.Stat(keys); err != nil || !info.IsDir() {
		return nil, nil
	}
	a, err := sshagent.Load(keys, "")
	if err != nil {
		return nil, fmt.Errorf("loading the keys in %s: %w", keys, err)
	}
	if a != nil {
		fmt.Fprintf(session, "\x1b[2m$ ssh-agent: %s from .ssh/\x1b[0m\r\n", strings.Join(a.Keys, ", "))
	}
	return a, nil
}
