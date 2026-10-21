package workspace

import (
	"path/filepath"

	"github.com/titpetric/task-ui/sshkey"
)

// InstanceDir returns the directory of the instance's own keys, beside
// the workspaces'.
func (k *Keyring) InstanceDir() string {
	return filepath.Join(k.Data, "instance")
}

// Instance returns the store of the instance's encryption key, the one
// that encrypts the services' secrets, such as the integrations'
// settings. No run, no git and no SSH connection gets it.
func (k *Keyring) Instance() *sshkey.Store {
	return &sshkey.Store{Dir: filepath.Join(k.InstanceDir(), "encrypt"), PublicMode: 0o600}
}

// InstanceKey returns the instance's encryption key, its public half
// and fingerprint; nil before EnsureInstance made it.
func (k *Keyring) InstanceKey() *sshkey.Key {
	for _, key := range k.Instance().List() {
		if !key.Backup {
			return &key
		}
	}
	return nil
}

// EnsureInstance makes the instance's key directory and its encryption
// key when it has none. It reports whether it generated the key; one
// there already stays.
func (k *Keyring) EnsureInstance(comment string) (bool, error) {
	if err := private(k.InstanceDir(), k.Instance().Dir); err != nil {
		return false, err
	}
	if k.InstanceKey() != nil {
		return false, nil
	}
	if _, err := k.Instance().Generate(encryptKind, comment); err != nil {
		return false, err
	}
	return true, nil
}

// RotateInstance generates a new encryption key for the instance; the
// previous one stays beside it as a backup, which still opens what it
// encrypted until a rekey writes the secrets again.
func (k *Keyring) RotateInstance(comment string) (sshkey.Key, error) {
	if err := private(k.InstanceDir(), k.Instance().Dir); err != nil {
		return sshkey.Key{}, err
	}
	return k.Instance().Generate(encryptKind, comment)
}
