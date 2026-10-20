// Package secret keeps the secrets encrypted at rest with the instance
// SSH keys: the public keys encrypt, the private key files decrypt. A
// rotation keeps the old pair as a backup, so what it encrypted still
// opens, and a rekey writes everything to the current keys again.
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

// ErrNoKey reports that no key in the SSH folder can encrypt.
var ErrNoKey = errors.New("no SSH key to encrypt the secrets with: generate an ed25519 or RSA key in Settings, SSH keys")

// Vault encrypts and decrypts with the keys of the SSH folder.
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

// Recipients returns the current keys that can encrypt, by name, leaving
// out the backups and the names excluded. ECDSA keys cannot encrypt.
func (v *Vault) Recipients(exclude ...string) []string {
	var names []string
	for _, key := range v.Keys.List() {
		if key.Backup || contains(exclude, key.Name) {
			continue
		}
		if _, err := recipient(key); err == nil {
			names = append(names, key.Name)
		}
	}
	return names
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

// identities returns every private key in the folder that can decrypt,
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

// Encrypt writes a value for the current keys, the excluded names left
// out, as armored text.
func (v *Vault) Encrypt(plain string, exclude ...string) (string, error) {
	var recipients []age.Recipient
	for _, key := range v.Keys.List() {
		if key.Backup || contains(exclude, key.Name) {
			continue
		}
		if r, err := recipient(key); err == nil {
			recipients = append(recipients, r)
		}
	}
	if len(recipients) == 0 {
		return "", ErrNoKey
	}
	var buf bytes.Buffer
	arm := armor.NewWriter(&buf)
	w, err := age.Encrypt(arm, recipients...)
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
		return "", errors.New("no SSH key opens the secret")
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
