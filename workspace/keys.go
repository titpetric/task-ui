// Package workspace keeps what a workspace holds beside its checkout:
// its keys, under the data directory, out of the checkout, where one
// git add would publish them, and out of the database.
package workspace

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/sshkey"
)

// Keyring keeps the keys of every workspace, a directory per workspace
// ID under Data/workspaces: encrypt/ holds the ed25519 key that
// encrypts its secrets, which no run gets, and ssh/ the SSH keys its
// runs get through their ssh agent. The directories are 0700, the key files
// 0600.
type Keyring struct {
	// Data is the data directory, TASK_UI_DATA.
	Data string
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
	for _, dir := range []string{k.Dir(id), k.Encrypt(id).Dir, k.SSH(id).Dir} {
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
// encryption key; the previous one stays beside it as a backup.
func (k *Keyring) UploadEncrypt(id string, data []byte, comment string) (sshkey.Key, error) {
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
