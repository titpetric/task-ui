// Package workspace keeps what a workspace holds beside its checkout:
// its keys, under the data directory, out of the checkout, where one
// git add would publish them, and out of the database. The instance's
// encryption key lives beside them.
package workspace

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/sshkey"
)

// Keyring keeps the keys of every workspace under Data/workspaces/<id>:
// encrypt/ holds the ed25519 key of its secrets, ssh/ the SSH keys of
// its runs. Data/instance/encrypt holds the integrations' key. A key
// has one use, encryption or SSH, and the keyring refuses it for the
// other. The directories are 0700, the key files 0600.
type Keyring struct {
	// Data is the data directory, TASK_UI_DATA.
	Data string
	// Access are the stores of the instance's SSH keys, outside the
	// keyring: TASK_UI_SSH_DIR and the server user's ~/.ssh.
	Access []*sshkey.Store
}

// encryptKind is the key type a workspace encrypts with.
const encryptKind = "ed25519"

// Dir returns the directory of a workspace's keys.
func (k *Keyring) Dir(id string) string {
	return filepath.Join(k.Data, "workspaces", id)
}

// Encrypt returns the store of a workspace's encryption key.
func (k *Keyring) Encrypt(id string) *sshkey.Store {
	return &sshkey.Store{Dir: filepath.Join(k.Dir(id), "encrypt"), PublicMode: 0o600}
}

// SSH returns the store of a workspace's SSH keys.
func (k *Keyring) SSH(id string) *sshkey.Store {
	return &sshkey.Store{Dir: filepath.Join(k.Dir(id), "ssh"), PublicMode: 0o600}
}

// EncryptKey returns a workspace's encryption key, its public half and
// fingerprint; nil for a workspace that has none yet.
func (k *Keyring) EncryptKey(id string) *sshkey.Key {
	for _, key := range k.Encrypt(id).List() {
		if !key.Backup {
			return &key
		}
	}
	return nil
}

// Prepare makes a workspace's key directories, private to the server,
// and no key: an admin adds the encryption key, generated or uploaded,
// before the workspace stores a secret.
func (k *Keyring) Prepare(id string) error {
	return k.dirs(id)
}

// Ensure makes a workspace's key directories and its encryption key
// when it has none. It reports whether it generated the key; one there
// already stays.
func (k *Keyring) Ensure(id, comment string) (bool, error) {
	if err := k.dirs(id); err != nil {
		return false, err
	}
	if k.EncryptKey(id) != nil {
		return false, nil
	}
	if _, err := k.Encrypt(id).Generate(encryptKind, comment); err != nil {
		return false, err
	}
	return true, nil
}

// dirs makes a workspace's key directories, private to the server.
func (k *Keyring) dirs(id string) error {
	if err := validID(id); err != nil {
		return err
	}
	return private(k.Dir(id), k.Encrypt(id).Dir, k.SSH(id).Dir)
}

// private makes directories, 0700.
func private(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// validID refuses an ID that would leave the keyring's directory.
func validID(id string) error {
	if id == "" || strings.ContainsAny(id, `/\.`) {
		return fmt.Errorf("invalid workspace id %q", id)
	}
	return nil
}

// RotateEncrypt generates a new encryption key for a workspace; the
// previous one stays beside it as a backup.
func (k *Keyring) RotateEncrypt(id, comment string) (sshkey.Key, error) {
	if err := k.dirs(id); err != nil {
		return sshkey.Key{}, err
	}
	return k.Encrypt(id).Generate(encryptKind, comment)
}

// ErrEncryptKind reports an uploaded encryption key that is not ed25519.
var ErrEncryptKind = errors.New("a workspace's encryption key is ed25519")

// UploadEncrypt stores an ed25519 private key as a workspace's
// encryption key; the previous one stays beside it as a backup. It
// refuses a key that is an SSH key anywhere or another's encryption
// key, ErrRole.
func (k *Keyring) UploadEncrypt(id string, data []byte, comment string) (sshkey.Key, error) {
	if err := k.CheckEncrypt(id, data); err != nil {
		return sshkey.Key{}, err
	}
	if private, err := ssh.ParseRawPrivateKey(data); err == nil {
		switch private.(type) {
		case ed25519.PrivateKey, *ed25519.PrivateKey:
		default:
			return sshkey.Key{}, ErrEncryptKind
		}
	}
	if err := k.dirs(id); err != nil {
		return sshkey.Key{}, err
	}
	return k.Encrypt(id).Upload(data, comment)
}

// Remove deletes a workspace's keys, when the workspace goes.
func (k *Keyring) Remove(id string) error {
	if err := validID(id); err != nil {
		return err
	}
	return os.RemoveAll(k.Dir(id))
}

// Files returns every file of a workspace's keys, backups among them,
// by its path in the workspace's directory: encrypt/<name> and
// ssh/<name>. An export bundles them.
func (k *Keyring) Files(id string) (map[string][]byte, error) {
	if err := validID(id); err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for _, sub := range []string{"encrypt", "ssh"} {
		entries, err := os.ReadDir(filepath.Join(k.Dir(id), sub))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(k.Dir(id), sub, entry.Name()))
			if err != nil {
				return nil, err
			}
			files[sub+"/"+entry.Name()] = data
		}
	}
	return files, nil
}

// Restore writes the key files of a bundle as a workspace's keys,
// 0600, into its private directories; Files names them. A workspace
// that has keys already keeps them: Restore refuses it. It refuses a
// bundle whose keys break their one use, ErrRole: an encryption key
// another workspace or the instance holds, or that is an SSH key
// anywhere, the bundle's own among them, and an SSH key that encrypts.
func (k *Keyring) Restore(id string, files map[string][]byte) error {
	if err := validID(id); err != nil {
		return err
	}
	if _, err := os.Stat(k.Dir(id)); err == nil {
		return fmt.Errorf("workspace %s has keys already", id)
	}
	if err := k.CheckBundle(files); err != nil {
		return err
	}
	if err := k.dirs(id); err != nil {
		return err
	}
	for name, data := range files {
		sub, base, ok := strings.Cut(name, "/")
		if !ok || (sub != "encrypt" && sub != "ssh") || base == "" || strings.ContainsAny(base, `/\`) || strings.HasPrefix(base, ".") {
			return fmt.Errorf("invalid key file %q", name)
		}
		if err := os.WriteFile(filepath.Join(k.Dir(id), sub, base), data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// Encrypts reports whether the key in a file is an encryption key: one
// the instance or a workspace encrypts with, wherever the file is. A
// run's ssh agent and git leave such a key out.
func (k *Keyring) Encrypts(path string) bool {
	fp, err := sshkey.FileFingerprint(path)
	if err != nil {
		return false
	}
	return len(k.encryption()[fp]) > 0
}

// Overlaps returns the keys held for both uses, by fingerprint: in an
// encryption store and in a store of SSH keys. Every file stays; an
// admin decides which copy goes.
func (k *Keyring) Overlaps() []Overlap {
	access := paths(k.accessStores()...)
	var result []Overlap
	for fp, encrypt := range k.encryption() {
		if held := access[fp]; len(held) > 0 {
			sort.Strings(encrypt)
			sort.Strings(held)
			result = append(result, Overlap{Fingerprint: fp, Encrypt: encrypt, Access: held})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Fingerprint < result[j].Fingerprint })
	return result
}
