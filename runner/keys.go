package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshagent"
)

// agent loads the keys a run gets into an ssh agent for it, and says so
// in the terminal: the instance keys the policies give whoever started
// the run, by their role or by name, and nothing else. A checkout's
// .ssh/ folder is not loaded. No key is no agent and no error, told as
// (none); the caller closes the agent.
func (m *Manager) agent(ctx context.Context, session *Session) (*sshagent.Agent, error) {
	var paths []string
	from := map[string]string{}
	if m.opts.KeyPolicies != nil && m.opts.KeyDir != "" && (session.role != "" || session.Run.UserName != "") {
		names, err := m.opts.KeyPolicies.For(ctx, session.role, session.Run.UserName)
		if err != nil {
			return nil, fmt.Errorf("reading the key policies: %w", err)
		}
		for _, name := range names {
			path := filepath.Join(m.opts.KeyDir, name)
			paths = append(paths, path)
			from[path] = name
		}
	}
	var a *sshagent.Agent
	if len(paths) > 0 {
		var err error
		a, err = sshagent.LoadFiles("", paths...)
		if err != nil {
			return nil, fmt.Errorf("loading the run's ssh keys: %w", err)
		}
	}
	labels := []string{"(none)"}
	if a != nil {
		labels = labels[:0]
		for _, path := range a.Keys {
			labels = append(labels, from[path])
		}
	}
	fmt.Fprintf(session, "\x1b[2m$ ssh-agent: %s\x1b[0m\r\n", strings.Join(labels, ", "))
	return a, nil
}

// gitSSH is the ssh command git in a job uses: the run's agent keys and
// no key file, ~/.ssh/id_* included, and the instance known_hosts.
func gitSSH(keyDir string) string {
	cmd := "ssh -o IdentityFile=none"
	if keyDir != "" {
		known := filepath.Join(keyDir, "known_hosts")
		cmd += " -o UserKnownHostsFile='" + strings.ReplaceAll(known, "'", `'\''`) + "'"
	}
	return cmd
}
