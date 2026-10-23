package keystore

import (
	"context"
)

// Restore takes the SSH keys of a bundle of the workspace source into a
// workspace's scope under folder, with the links the bundle names for
// them by their names, and returns the names it took. A name the
// folder holds already refuses the key, ErrExists.
func (l *Local) Restore(ctx context.Context, scope, source, folder string, files map[string][]byte, roles, users map[string][]string) ([]string, error) {
	keys, err := unbundle(files, source, folder)
	if err != nil {
		return nil, err
	}
	var taken []string
	for _, b := range keys {
		key, err := l.Import(ctx, Spec{Name: b.name, Purpose: PurposeSSH, Scope: scope}, b.data)
		if err != nil {
			return taken, err
		}
		taken = append(taken, key.Name)
		if len(roles[b.from])+len(users[b.from]) > 0 {
			if err := l.Link(ctx, key.Name, roles[b.from], users[b.from]); err != nil {
				return taken, err
			}
		}
	}
	return taken, nil
}
