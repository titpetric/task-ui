package sshkey

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Types are the key types the store generates, by file name.
var Types = map[string]string{
	"ed25519": "id_ed25519",
	"ecdsa":   "id_ecdsa",
	"rsa":     "id_rsa",
}

// Store manages the SSH key pairs in one folder, usually $HOME/.ssh. A key
// pair is the private key file and its .pub sibling. Replacing a key moves
// the previous pair aside with a timestamped .bak suffix, so a rotation can
// be undone by hand.
type Store struct {
	Dir string
}

var keyName = regexp.MustCompile(`^id_[a-z0-9]+(\.bak-[0-9]{8}T[0-9]{6})?$`)

// ErrPassphrase is returned for an uploaded key protected by a passphrase:
// the jobs run unattended and could not unlock it.
var ErrPassphrase = errors.New("the key is protected by a passphrase, upload an unencrypted key")

// List returns the key pairs and their backups.
func (k *Store) List() []Key {
	entries, _ := os.ReadDir(k.Dir)
	result := []Key{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !keyName.MatchString(name) {
			continue
		}
		key, err := k.describe(name)
		if err != nil {
			continue
		}
		result = append(result, key)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Backup != result[j].Backup {
			return !result[i].Backup
		}
		return result[i].Name > result[j].Name
	})
	return result
}

func (k *Store) describe(name string) (Key, error) {
	path := filepath.Join(k.Dir, name)
	info, err := os.Stat(path)
	if err != nil {
		return Key{}, err
	}

	var (
		pub     ssh.PublicKey
		comment string
	)
	if data, err := os.ReadFile(path + ".pub"); err == nil {
		pub, comment, _, _, err = ssh.ParseAuthorizedKey(data)
		if err != nil {
			pub = nil
		}
	}
	if pub == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return Key{}, err
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return Key{}, err
		}
		pub = signer.PublicKey()
	}

	return Key{
		Name:        name,
		Type:        pub.Type(),
		Bits:        keyBits(pub),
		Fingerprint: ssh.FingerprintSHA256(pub),
		Comment:     comment,
		PublicKey:   strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub))) + commentSuffix(comment),
		Modified:    info.ModTime(),
		Backup:      strings.Contains(name, ".bak-"),
	}, nil
}

func commentSuffix(comment string) string {
	if comment == "" {
		return ""
	}
	return " " + comment
}

func keyBits(pub ssh.PublicKey) int {
	crypto, ok := pub.(ssh.CryptoPublicKey)
	if !ok {
		return 0
	}
	switch key := crypto.CryptoPublicKey().(type) {
	case *rsa.PublicKey:
		return key.N.BitLen()
	case *ecdsa.PublicKey:
		return key.Curve.Params().BitSize
	case ed25519.PublicKey:
		return 256
	}
	return 0
}

// Generate creates a new key pair of the given type, moving an existing pair
// of the same type aside. It returns the new key.
func (k *Store) Generate(kind, comment string) (Key, error) {
	name, ok := Types[kind]
	if !ok {
		return Key{}, fmt.Errorf("unsupported key type %q", kind)
	}

	var private any
	switch kind {
	case "ed25519":
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return Key{}, err
		}
		private = key
	case "ecdsa":
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return Key{}, err
		}
		private = key
	case "rsa":
		key, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return Key{}, err
		}
		private = key
	}

	block, err := ssh.MarshalPrivateKey(private, comment)
	if err != nil {
		return Key{}, err
	}
	return k.write(name, pem.EncodeToMemory(block), private, comment)
}

// Upload stores a private key given in PEM or OpenSSH format, moving an
// existing pair of the same type aside.
func (k *Store) Upload(data []byte, comment string) (Key, error) {
	data = []byte(strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n")) + "\n")
	private, err := ssh.ParseRawPrivateKey(data)
	if err != nil {
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			return Key{}, ErrPassphrase
		}
		return Key{}, fmt.Errorf("not a private key: %w", err)
	}

	if key, ok := private.(*ed25519.PrivateKey); ok {
		private = *key
	}

	var name string
	switch private.(type) {
	case ed25519.PrivateKey:
		name = Types["ed25519"]
	case *ecdsa.PrivateKey:
		name = Types["ecdsa"]
	case *rsa.PrivateKey:
		name = Types["rsa"]
	default:
		return Key{}, fmt.Errorf("unsupported key type %T", private)
	}
	return k.write(name, data, private, comment)
}

func (k *Store) write(name string, privatePEM []byte, private any, comment string) (Key, error) {
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return Key{}, err
	}
	if err := os.MkdirAll(k.Dir, 0o700); err != nil {
		return Key{}, err
	}
	if err := k.backup(name); err != nil {
		return Key{}, err
	}

	path := filepath.Join(k.Dir, name)
	if err := os.WriteFile(path, privatePEM, 0o600); err != nil {
		return Key{}, err
	}
	public := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))) + commentSuffix(comment) + "\n"
	if err := os.WriteFile(path+".pub", []byte(public), 0o644); err != nil {
		return Key{}, err
	}
	return k.describe(name)
}

// backup moves an existing key pair aside.
func (k *Store) backup(name string) error {
	path := filepath.Join(k.Dir, name)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	suffix := ".bak-" + time.Now().UTC().Format("20060102T150405")
	if err := os.Rename(path, path+suffix); err != nil {
		return err
	}
	if _, err := os.Stat(path + ".pub"); err == nil {
		return os.Rename(path+".pub", path+suffix+".pub")
	}
	return nil
}

// Delete removes a key pair or a backup.
func (k *Store) Delete(name string) error {
	if !keyName.MatchString(name) {
		return fmt.Errorf("invalid key name %q", name)
	}
	path := filepath.Join(k.Dir, name)
	if err := os.Remove(path); err != nil {
		return err
	}
	if err := os.Remove(path + ".pub"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
