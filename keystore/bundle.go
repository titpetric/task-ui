package keystore

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/sshkey"
)

// Bundle returns the SSH keys of a scope as a workspace export bundles
// them, by path, ssh/<name>.key, and the roles and users of each by
// name. No encryption key leaves: the instance's encrypts every secret,
// and an export carries the secrets inside its passphrase.
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
		files[key.Purpose+"/"+key.Name+".key"] = data
		if len(key.Roles) > 0 {
			roles[key.Name] = key.Roles
		}
		if len(key.Users) > 0 {
			users[key.Name] = key.Users
		}
	}
	return files, roles, users, nil
}

// bundled is a private key of a bundle: the name it had, the name it
// takes, its purpose and data.
type bundled struct {
	from, name, purpose string
	data                []byte
}

// unbundle reads the private keys of a bundle, public halves and
// backups left out. A workspace's key of the source folder,
// workspace_<source>_..., takes the folder's name in its place, and a
// key of an older bundle, id_<type>[_<label>], the name of a
// workspace's key under folder.
func unbundle(files map[string][]byte, source, folder string) ([]bundled, error) {
	var result []bundled
	for path, data := range files {
		dir, file, ok := strings.Cut(path, "/")
		purpose := map[string]string{"encrypt": PurposeEncrypt, "ssh": PurposeSSH}[dir]
		if !ok || purpose == "" || strings.ContainsAny(file, `/\`) {
			return nil, fmt.Errorf("invalid key file %q", path)
		}
		if strings.HasSuffix(file, ".pub") || strings.Contains(file, ".bak-") {
			continue
		}
		from := strings.TrimSuffix(file, ".key")
		name := from
		if rest, ok := strings.CutPrefix(name, WorkspaceKeyName(source, "")); ok && source != "" {
			name = WorkspaceKeyName(folder, "") + rest
		}
		if sshkey.KindOf(file) != "" {
			name = WorkspaceKeyName(folder, label(file, map[string]string{PurposeEncrypt: "encryption", PurposeSSH: "ssh"}[purpose]))
		}
		result = append(result, bundled{from: from, name: name, purpose: purpose, data: data})
	}
	return result, nil
}

// CheckBundle refuses the keys of a bundle the folder holds already,
// under any name or purpose, ErrExists: an import checks before it
// clones anything.
func (l *Local) CheckBundle(ctx context.Context, files map[string][]byte) error {
	keys, err := unbundle(files, "", "x")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if key.purpose == PurposeEncrypt {
			continue
		}
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

// BundledEncryptKeys returns the private encryption keys an older
// bundle carries, encrypt/<name>.key, which open the secrets it
// carries encrypted.
func BundledEncryptKeys(files map[string][]byte) [][]byte {
	var result [][]byte
	keys, err := unbundle(files, "", "x")
	if err != nil {
		return nil
	}
	for _, key := range keys {
		if key.purpose == PurposeEncrypt {
			result = append(result, key.data)
		}
	}
	return result
}
