// Package secret keeps the secrets encrypted at rest with age, to
// ed25519 or RSA keys kept for nothing else: the instance's encryption
// key for the services' secrets, each workspace's for its own. The
// public key encrypts, the private key file decrypts. A rotation keeps
// the old pair as a backup, so what it encrypted still opens, and a
// rekey writes everything to the current keys again. No SSH key, the
// ones runs and git authenticate with, encrypts or decrypts.
package secret

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"filippo.io/age/armor"
	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/sshkey"
)

// ErrNoKey reports that no key in the vault's store can encrypt.
var ErrNoKey = errors.New("no encryption key to encrypt the secrets with: rotate it in Settings, SSH keys, Encryption")

// Vault encrypts and decrypts with the keys of one encryption store.
type Vault struct {
	Keys *sshkey.Store
}

// NewVault returns a vault over the key store.
func NewVault(keys *sshkey.Store) *Vault {
	return &Vault{Keys: keys}
}

// Encrypted reports whether a value is one the vault wrote.
func Encrypted(text string) bool {
	return strings.HasPrefix(text, armor.Header)
}

// Recipients returns the store's key, by name, the one key that
// encrypts the secrets, or none; it skips the names excluded.
func (v *Vault) Recipients(exclude ...string) []string {
	if key, _, ok := v.instanceKey(exclude); ok {
		return []string{key.Name}
	}
	return nil
}

// instanceKey picks the key that encrypts the secrets: a type's own
// key, id_ed25519 before id_rsa, never a backup and never a key made for
// someone or for a use, such as id_rsa_deploy. ECDSA keys cannot encrypt.
func (v *Vault) instanceKey(exclude []string) (sshkey.Key, age.Recipient, bool) {
	for _, kind := range []string{"ed25519", "rsa"} {
		for _, key := range v.Keys.List() {
			if key.Backup || key.Label != "" || sshkey.KindOf(key.Name) != kind || contains(exclude, key.Name) {
				continue
			}
			if r, err := recipient(key); err == nil {
				return key, r, true
			}
		}
	}
	return sshkey.Key{}, nil, false
}

// recipient turns a key into an age recipient.
func recipient(key sshkey.Key) (age.Recipient, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.PublicKey))
	if err != nil {
		return nil, err
	}
	switch pub.Type() {
	case ssh.KeyAlgoED25519:
		return agessh.NewEd25519Recipient(pub)
	case ssh.KeyAlgoRSA:
		return agessh.NewRSARecipient(pub)
	}
	return nil, fmt.Errorf("%s keys cannot encrypt", pub.Type())
}

// identities returns every private key in the store that can decrypt,
// backups included.
func (v *Vault) identities() []age.Identity {
	var result []age.Identity
	for _, key := range v.Keys.List() {
		data, err := os.ReadFile(filepath.Join(v.Keys.Dir, key.Name))
		if err != nil {
			continue
		}
		if identity, err := agessh.ParseIdentity(data); err == nil {
			result = append(result, identity)
		}
	}
	return result
}

// Encrypt writes a value for the store's key, the excluded names
// passed over, as armored text.
func (v *Vault) Encrypt(plain string, exclude ...string) (string, error) {
	_, r, ok := v.instanceKey(exclude)
	if !ok {
		return "", ErrNoKey
	}
	var buf bytes.Buffer
	arm := armor.NewWriter(&buf)
	w, err := age.Encrypt(arm, r)
	if err != nil {
		return "", err
	}
	if _, err := io.WriteString(w, plain); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	if err := arm.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Decrypt reads a value with whichever key opens it. It returns a value
// without encryption unchanged.
func (v *Vault) Decrypt(text string) (string, error) {
	if !Encrypted(text) {
		return text, nil
	}
	identities := v.identities()
	if len(identities) == 0 {
		return "", errors.New("no key opens the secret")
	}
	r, err := age.Decrypt(armor.NewReader(strings.NewReader(text)), identities...)
	if err != nil {
		return "", fmt.Errorf("opening the secret: %w", err)
	}
	plain, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
