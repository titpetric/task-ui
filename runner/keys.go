package runner

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/sshagent"
)

// agent loads the SSH keys a run gets into an ssh agent of its own,
// from memory, and says so in the terminal: the instance's keys linked
// to the role or the name of whoever started the run, and the keys of
// its workspace with no link or linked so; a host command gets every
// instance SSH key, which the inventory signs in with. No encryption key serves a
// run. No key is no agent and no error, told as (none); the caller
// closes the agent, which removes its socket.
func (m *Manager) agent(ctx context.Context, session *Session) (*sshagent.Agent, error) {
	if m.opts.Keys == nil {
		fmt.Fprintf(session, "\x1b[2m$ ssh-agent: (none)\x1b[0m\r\n")
		return nil, nil
	}
	names, err := m.opts.Keys.ForRun(ctx, session.role, session.Run.UserName, session.workspaceID)
	if err != nil {
		return nil, fmt.Errorf("reading the run's keys: %w", err)
	}
	if session.Job.IsHost() {
		// A host command signs in with the instance's keys, as the
		// inventory's probe does.
		logins, err := m.opts.Keys.Logins(ctx)
		if err != nil {
			return nil, err
		}
		for _, name := range logins {
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
	}
	a, err := m.opts.Keys.Materialize(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("loading the run's ssh keys: %w", err)
	}
	labels := []string{"(none)"}
	if a != nil {
		labels = a.Keys
	}
	fmt.Fprintf(session, "\x1b[2m$ ssh-agent: %s\x1b[0m\r\n", strings.Join(labels, ", "))
	return a, nil
}
