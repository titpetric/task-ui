package keystore

import (
	"context"

	"github.com/titpetric/task-ui/sshagent"
)

// gitName returns the name of the key git uses, "" for none.
func (l *Local) gitName(ctx context.Context) string {
	if l.Settings == nil {
		return ""
	}
	values, err := l.Settings.Values(ctx)
	if err != nil {
		return ""
	}
	return values[gitSetting]
}

// Git returns the key git pulls and clones with, nil for none.
func (l *Local) Git(ctx context.Context) *Key {
	name := l.gitName(ctx)
	if name == "" {
		return nil
	}
	key, err := l.Get(ctx, name)
	if err != nil {
		return nil
	}
	return key
}

// SetGit makes an SSH key the one git pulls and clones with; "" none.
func (l *Local) SetGit(ctx context.Context, name string) error {
	if name != "" {
		row, err := l.Rows.Get(ctx, name)
		if err != nil {
			return err
		}
		if row.Purpose != PurposeSSH {
			return ErrNotSSH
		}
	}
	return l.Settings.Set(ctx, map[string]string{gitSetting: name})
}

// GitAgent serves the git key in an ssh agent, with the path of its
// public half for ssh -i; a nil agent with no git key.
func (l *Local) GitAgent(ctx context.Context) (*sshagent.Agent, string, error) {
	key := l.Git(ctx)
	if key == nil {
		return nil, "", nil
	}
	agent, err := l.Materialize(ctx, []string{key.Name})
	return agent, l.PublicPath(key.Name), err
}
