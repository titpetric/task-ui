package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshagent"
)

// agent loads the keys a run gets into an ssh agent for it, and says so
// in the terminal: the private keys in the .ssh/ folder of the checkout
// the folder is in, and the instance keys the policies give the role of
// whoever started the run. No key is no agent and no error; the caller
// closes the agent.
func (m *Manager) agent(ctx context.Context, session *Session) (*sshagent.Agent, error) {
	var paths []string
	from := map[string]string{}
	if dir := m.checkout(session); dir != "" {
		keys := filepath.Join(dir, ".ssh")
		if entries, err := os.ReadDir(keys); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					path := filepath.Join(keys, entry.Name())
					paths = append(paths, path)
					from[path] = ".ssh/" + entry.Name()
				}
			}
		}
	}
	if m.opts.KeyPolicies != nil && m.opts.KeyDir != "" && session.role != "" {
		names, err := m.opts.KeyPolicies.Keys(ctx, session.role)
		if err != nil {
			return nil, fmt.Errorf("reading the key policies: %w", err)
		}
		for _, name := range names {
			path := filepath.Join(m.opts.KeyDir, name)
			paths = append(paths, path)
			from[path] = name + " (" + session.role + ")"
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}
	a, err := sshagent.LoadFiles("", paths...)
	if err != nil {
		return nil, fmt.Errorf("loading the run's ssh keys: %w", err)
	}
	if a != nil {
		labels := make([]string, 0, len(a.Keys))
		for _, path := range a.Keys {
			labels = append(labels, from[path])
		}
		fmt.Fprintf(session, "\x1b[2m$ ssh-agent: %s\x1b[0m\r\n", strings.Join(labels, ", "))
	}
	return a, nil
}
