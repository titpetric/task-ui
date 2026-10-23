package keystore

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/titpetric/task-ui/sshkey"
)

// Bundle returns the SSH keys of a scope as a workspace export bundles
// them, by file name, <name>.key and <name>.pub, and the roles and
// users of each by name. These are the keys the workspace's runs get;
// no instance key and no encryption key leaves: the instance's encrypts
// every secret, and an export carries the secrets inside its
// passphrase.
func (l *Local) Bundle(ctx context.Context, scope string) (map[string][]byte, map[string][]string, map[string][]string, error) {
	keys, err := l.Of(ctx, scope, PurposeSSH)
	if err != nil {
		return nil, nil, nil, err
	}
	files, roles, users := map[string][]byte{}, map[string][]string{}, map[string][]string{}
	for _, key := range keys {
		data, err := l.Private(ctx, key.Name)
		if err != nil {
			return nil, nil, nil, err
		}
		files[key.Name+".key"] = data
		if public, err := os.ReadFile(l.PublicPath(key.Name)); err == nil {
			files[key.Name+".pub"] = public
		}
		if len(key.Roles) > 0 {
			roles[key.Name] = key.Roles
		}
		if len(key.Users) > 0 {
			users[key.Name] = key.Users
		}
	}
	return files, roles, users, nil
}

// bundled is a private SSH key of a bundle: the name it had, the name
// it takes and its data.
type bundled struct {
	from, name string
	data       []byte
}

// unbundle reads the private keys of a bundle, <name>.key, the public
// halves left out: the import derives them. A workspace's key of the
// source folder, workspace_<source>_..., takes the folder's name in
// its place.
func unbundle(files map[string][]byte, source, folder string) ([]bundled, error) {
	var result []bundled
	for file, data := range files {
		if strings.ContainsAny(file, `/\`) {
			return nil, fmt.Errorf("invalid key file %q", file)
		}
		if strings.HasSuffix(file, ".pub") {
			continue
		}
		from, ok := strings.CutSuffix(file, ".key")
		if !ok {
			return nil, fmt.Errorf("invalid key file %q", file)
		}
		name := from
		if rest, ok := strings.CutPrefix(name, WorkspaceKeyName(source, "")); ok && source != "" {
			name = WorkspaceKeyName(folder, "") + rest
		}
		result = append(result, bundled{from: from, name: name, data: data})
	}
	return result, nil
}

// CheckBundle refuses the keys of a bundle the folder holds already,
// under any name, ErrExists: an import checks before it clones
// anything.
func (l *Local) CheckBundle(ctx context.Context, files map[string][]byte) error {
	keys, err := unbundle(files, "", "x")
	if err != nil {
		return err
	}
	for _, key := range keys {
		fp, err := sshkey.Fingerprint(key.data)
		if err != nil {
			return fmt.Errorf("the bundle's key %s: %w", key.name, err)
		}
		if row, err := l.Rows.ByFingerprint(ctx, fp); err == nil {
			return fmt.Errorf("%w: the bundle's %s is %s (%s)", ErrExists, key.name, row.Name, fp)
		}
	}
	return nil
}
