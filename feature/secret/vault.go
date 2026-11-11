// Package secret keeps the secrets encrypted at rest with age, to
// ed25519 or RSA keys kept for nothing else: the instance's one
// encryption key for every secret, the workspaces' and the services'.
// The public key encrypts, the private key file decrypts. A rotation keeps
// the old pair as a backup, so what it encrypted still opens, and a
// rekey writes everything to the current keys again. No SSH key, the
// ones runs and git authenticate with, encrypts or decrypts.
package secret

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"filippo.io/age/armor"
	"golang.org/x/crypto/ssh"
)

// ErrNoKey reports that the vault has no key to encrypt with.
var ErrNoKey = errors.New("no encryption key to encrypt the secrets with: generate or upload it in Settings")

// Keys are what a vault encrypts and decrypts with: one encryption key,
// and the private halves that open what it and its earlier material
// encrypted.
type Keys interface {
	// Recipient returns the name and the authorized_keys line of the
	// key that encrypts; ok false for none.
	Recipient() (name, public string, ok bool)
	// Identities returns the private keys that decrypt, the current
	// first.
	Identities() [][]byte
}

// Vault encrypts and decrypts with one encryption key.
type Vault struct {
	Keys Keys
}

// NewVault returns a vault over the keys.
func NewVault(keys Keys) *Vault {
	return &Vault{Keys: keys}
}

// Encrypted reports whether a value is one the vault wrote.
func Encrypted(text string) bool {
	return strings.HasPrefix(text, armor.Header)
}

// Recipients returns the key, by name, that encrypts the secrets, or
// none; it skips the names excluded.
func (v *Vault) Recipients(exclude ...string) []string {
	if name, _, ok := v.recipient(exclude); ok {
		return []string{name}
	}
	return nil
}

// recipient returns the key that encrypts, as an age recipient.
func (v *Vault) recipient(exclude []string) (string, age.Recipient, bool) {
	if v == nil || v.Keys == nil {
		return "", nil, false
	}
	name, public, ok := v.Keys.Recipient()
	if !ok || contains(exclude, name) {
		return "", nil, false
	}
	r, err := recipient(public)
	if err != nil {
		return "", nil, false
	}
	return name, r, true
}

// recipient turns a public key into an age recipient.
func recipient(public string) (age.Recipient, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(public))
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

// identities returns every private key that can decrypt.
func (v *Vault) identities() []age.Identity {
	if v == nil || v.Keys == nil {
		return nil
	}
	var result []age.Identity
	for _, data := range v.Keys.Identities() {
		if identity, err := agessh.ParseIdentity(data); err == nil {
			result = append(result, identity)
		}
	}
	return result
}

// Encrypt writes a value for the store's key, the excluded names
// passed over, as armored text.
func (v *Vault) Encrypt(plain string, exclude ...string) (string, error) {
	_, r, ok := v.recipient(exclude)
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

// Decrypt reads a value with the key that opens it. It returns a value
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
