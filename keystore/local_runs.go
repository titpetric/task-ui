package keystore

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/titpetric/task-ui/sshagent"
	"github.com/titpetric/task-ui/storage"
)

// ForRun returns the SSH keys a run gets, by name: the instance's
// linked to the role or the name of whoever started it, and the run's
// workspace's with no link or linked so.
func (l *Local) ForRun(ctx context.Context, role, user, workspaceID string) ([]string, error) {
	keys, err := l.List(ctx)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, key := range keys {
		if key.Purpose != PurposeSSH {
			continue
		}
		linked := (role != "" && slices.Contains(key.Roles, role)) || (user != "" && slices.Contains(key.Users, user))
		switch key.Scope {
		case storage.ScopeInstance:
			if linked {
				names = append(names, key.Name)
			}
		case ScopeWorkspace(workspaceID):
			if workspaceID != "" && (linked || len(key.Roles)+len(key.Users) == 0) {
				names = append(names, key.Name)
			}
		}
	}
	return names, nil
}

// Materialize serves SSH keys in an ssh agent, from memory, on a socket
// in a 0700 directory under Tmp. An encryption key is ErrNotSSH. No key
// is a nil agent.
func (l *Local) Materialize(ctx context.Context, names []string) (*sshagent.Agent, error) {
	var keys []sshagent.Named
	for _, name := range names {
		row, err := l.Rows.Get(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("the key %s: %w", name, err)
		}
		if row.Purpose != PurposeSSH {
			return nil, fmt.Errorf("%s: %w", name, ErrNotSSH)
		}
		data, err := os.ReadFile(l.path(name, ".key"))
		if err != nil {
			return nil, err
		}
		keys = append(keys, sshagent.Named{Name: name, PEM: data})
	}
	return sshagent.LoadKeys(l.Tmp, keys...)
}

// Logins returns the instance's SSH keys, by name, the ones the
// inventory signs in to hosts with.
func (l *Local) Logins(ctx context.Context) ([]string, error) {
	keys, err := l.Of(ctx, storage.ScopeInstance, PurposeSSH)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(keys))
	for i := range keys {
		names[i] = keys[i].Name
	}
	return names, nil
}

// LoginAgent serves the instance's SSH keys in an ssh agent, for a
// probe of a host; nil with none.
func (l *Local) LoginAgent(ctx context.Context) (*sshagent.Agent, error) {
	names, err := l.Logins(ctx)
	if err != nil {
		return nil, err
	}
	return l.Materialize(ctx, names)
}
