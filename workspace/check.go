package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshkey"
)

// ErrRole reports a key that would serve both uses, encryption and SSH,
// or one encryption key for two holders.
var ErrRole = errors.New("a key either encrypts secrets or authenticates SSH connections, never both, and an encryption key has one holder")

// encryptStores returns every encryption store by its holder: the
// instance as "", each workspace by its ID.
func (k *Keyring) encryptStores() map[string]*sshkey.Store {
	stores := map[string]*sshkey.Store{"": k.Instance()}
	for _, id := range k.ids() {
		stores[id] = k.Encrypt(id)
	}
	return stores
}

// accessStores returns every store of SSH keys: the instance's and each
// workspace's ssh/.
func (k *Keyring) accessStores() []*sshkey.Store {
	stores := append([]*sshkey.Store{}, k.Access...)
	for _, id := range k.ids() {
		stores = append(stores, k.SSH(id))
	}
	return stores
}

// ids returns the IDs of the workspaces that have a key directory.
func (k *Keyring) ids() []string {
	entries, _ := os.ReadDir(filepath.Join(k.Data, "workspaces"))
	var ids []string
	for _, entry := range entries {
		if entry.IsDir() && validID(entry.Name()) == nil {
			ids = append(ids, entry.Name())
		}
	}
	return ids
}

// paths returns the key files of stores by fingerprint, backups among
// them: a backup still opens or reaches what its key did.
func paths(stores ...*sshkey.Store) map[string][]string {
	result := map[string][]string{}
	for _, store := range stores {
		if store == nil {
			continue
		}
		for _, key := range store.List() {
			result[key.Fingerprint] = append(result[key.Fingerprint], filepath.Join(store.Dir, key.Name))
		}
	}
	return result
}

// encryption returns the files of every encryption key by fingerprint.
func (k *Keyring) encryption() map[string][]string {
	var stores []*sshkey.Store
	for _, store := range k.encryptStores() {
		stores = append(stores, store)
	}
	return paths(stores...)
}

// CheckAccess refuses key material, a private or a public key, that an
// encryption store holds, ErrRole: it cannot become an SSH key.
func (k *Keyring) CheckAccess(data []byte) error {
	fp, err := sshkey.Fingerprint(data)
	if err != nil {
		return nil
	}
	if held := k.encryption()[fp]; len(held) > 0 {
		return fmt.Errorf("%w: %s encrypts with this key (%s)", ErrRole, held[0], fp)
	}
	return nil
}

// CheckEncrypt refuses key material as the encryption key of a holder,
// a workspace by ID or the instance as "", ErrRole, when an SSH store
// anywhere holds it or another holder encrypts with it. The holder's own
// backups may come back.
func (k *Keyring) CheckEncrypt(id string, data []byte) error {
	fp, err := sshkey.Fingerprint(data)
	if err != nil {
		return nil
	}
	return k.checkEncrypt(k.encryptDir(id), fp, nil)
}

// encryptDir returns the encryption store's directory of a holder, a
// workspace by ID or the instance as "".
func (k *Keyring) encryptDir(id string) string {
	if id == "" {
		return k.Instance().Dir
	}
	return k.Encrypt(id).Dir
}

// checkEncrypt refuses a fingerprint as the encryption key of the
// holder whose store is in own when an SSH store, or the extra
// fingerprints, hold it, or another holder encrypts with it. An empty
// own is a holder still to come, which holds nothing yet.
func (k *Keyring) checkEncrypt(own, fp string, extra map[string]string) error {
	if held := paths(k.accessStores()...)[fp]; len(held) > 0 {
		return fmt.Errorf("%w: %s is this key as an SSH key (%s)", ErrRole, held[0], fp)
	}
	if name, ok := extra[fp]; ok {
		return fmt.Errorf("%w: the bundle's %s is this key as an SSH key (%s)", ErrRole, name, fp)
	}
	for _, path := range k.encryption()[fp] {
		if filepath.Dir(path) != own {
			return fmt.Errorf("%w: %s encrypts with this key already (%s)", ErrRole, path, fp)
		}
	}
	return nil
}

// CheckBundle refuses the key files of a bundle, as Files names them,
// for a workspace still to come, ErrRole, when one breaks its use: an
// encryption key the instance or a workspace holds, or one that is an
// SSH key anywhere, the bundle's own SSH keys among them, or an SSH key
// that encrypts. Public halves count as their keys. An import checks
// before it clones anything; Restore checks again.
func (k *Keyring) CheckBundle(files map[string][]byte) error {
	encrypt, access := map[string]string{}, map[string]string{}
	for name, data := range files {
		fp, err := sshkey.Fingerprint(data)
		if err != nil {
			continue
		}
		if strings.HasPrefix(name, "encrypt/") {
			encrypt[fp] = name
		} else {
			access[fp] = name
		}
	}
	for fp := range encrypt {
		if err := k.checkEncrypt("", fp, access); err != nil {
			return err
		}
	}
	for fp, name := range access {
		if held := k.encryption()[fp]; len(held) > 0 {
			return fmt.Errorf("%w: the bundle's %s is %s, an encryption key (%s)", ErrRole, name, held[0], fp)
		}
	}
	return nil
}
