package keystore

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/titpetric/task-ui/sshagent"
	"github.com/titpetric/task-ui/storage"
)

// ForRun returns the names of the SSH keys a run gets: the instance's linked to the starter's role or name,
// and those of the run's workspace linked the same way or to nobody.
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

// Materialize serves SSH keys from memory in an ssh agent on a socket under Tmp, each key's backup after it.
// An encryption key is ErrNotSSH, and no key returns a nil agent.
func (l *Local) Materialize(ctx context.Context, names []string) (*sshagent.Agent, error) {
	keys, err := l.AgentKeys(ctx, names)
	if err != nil {
		return nil, err
	}
	return sshagent.LoadKeys(l.Tmp, keys...)
}

// AgentKeys returns the private halves an agent serves for SSH keys, in
// order, each key's backup right after it: a rotated key keeps a host
// that knows its previous public key alone letting the run in.
func (l *Local) AgentKeys(ctx context.Context, names []string) ([]sshagent.Named, error) {
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
		if row.Backup != "" {
			if data, err := os.ReadFile(l.backupPath(name, ".key")); err == nil {
				keys = append(keys, sshagent.Named{Name: row.Backup, PEM: data})
			}
		}
	}
	return keys, nil
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
