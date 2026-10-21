package runner

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshagent"
)

// agent loads the keys a run gets into an ssh agent for it, and says so
// in the terminal: the instance keys the policies grant whoever started
// the run, by their role or by name, then the SSH keys of the run's
// workspace its grants let through, each as ws:<name>. It loads no
// encryption key, the instance's or a workspace's, nor a copy of one in
// an SSH folder, and no key of a checkout's .ssh/ folder. No key is no
// agent and no error, told as (none); the caller closes the agent.
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
			if m.encrypts(path) {
				continue
			}
			paths = append(paths, path)
			from[path] = name
		}
	}
	more, err := m.workspaceKeys(ctx, session)
	if err != nil {
		return nil, err
	}
	for _, name := range more {
		path := filepath.Join(m.opts.WorkspaceKeys.SSH(session.workspaceID).Dir, name)
		if m.encrypts(path) {
			continue
		}
		paths = append(paths, path)
		from[path] = "ws:" + name
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

// workspaceKeys returns the names of the SSH keys of the run's workspace
// the run gets: every current key without a grant, and those granted to
// the role or the name of whoever started it.
func (m *Manager) workspaceKeys(ctx context.Context, session *Session) ([]string, error) {
	if m.opts.WorkspaceKeys == nil || session.workspaceID == "" {
		return nil, nil
	}
	var names []string
	for _, key := range m.opts.WorkspaceKeys.SSH(session.workspaceID).List() {
		if !key.Backup {
			names = append(names, key.Name)
		}
	}
	if m.opts.KeyGrants == nil || len(names) == 0 {
		return names, nil
	}
	names, err := m.opts.KeyGrants.For(ctx, session.workspaceID, names, session.role, session.Run.UserName)
	if err != nil {
		return nil, fmt.Errorf("reading the workspace's key grants: %w", err)
	}
	return names, nil
}

// encrypts reports whether a key file is an encryption key, which no
// run gets; it logs the one it withholds.
func (m *Manager) encrypts(path string) bool {
	if m.opts.WorkspaceKeys == nil || !m.opts.WorkspaceKeys.Encrypts(path) {
		return false
	}
	slog.Warn("an encryption key in an SSH folder is not loaded into the run's agent", "key", path)
	return true
}
