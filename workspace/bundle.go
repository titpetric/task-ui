package workspace

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
)

// Extension ends the file name of a bundle: <folder>.tui.age, an age
// file.
const Extension = ".tui.age"

// MinPassphrase is the length a passphrase an export encrypts to takes
// at least.
const MinPassphrase = 12

// The limits ReadBundle holds a bundle to: the encrypted file, the manifest,
// each key file and the number of entries.
const (
	MaxBundle   = 64 << 20
	MaxManifest = 16 << 20
	MaxKeyFile  = 1 << 20
	MaxEntries  = 1024
)

// manifestName and sshDir are where a bundle keeps the manifest and
// the SSH key files.
const (
	manifestName = "manifest.json"
	sshDir       = ".ssh/"
)

// ErrPassphrase reports a passphrase that does not open a bundle.
var ErrPassphrase = errors.New("the passphrase does not open the bundle")

// ErrBundle reports a file that is not a bundle, or one that breaks its
// limits.
var ErrBundle = errors.New("not a workspace bundle")

// ErrWeakPassphrase refuses a passphrase shorter than MinPassphrase.
var ErrWeakPassphrase = fmt.Errorf("a passphrase takes at least %d characters", MinPassphrase)

// CheckPassphrase refuses a passphrase too short to encrypt a bundle
// to.
func CheckPassphrase(passphrase string) error {
	if len([]rune(passphrase)) < MinPassphrase {
		return ErrWeakPassphrase
	}
	return nil
}

// Bundle is a workspace as one instance exports it for another: a tar.gz
// of manifest.json and .ssh/, the SSH keys of the workspace, encrypted
// to a passphrase with age's scrypt. The secrets in the manifest are
// plain inside it; nothing else encrypts them.
type Bundle struct {
	Manifest Manifest
	// Keys are the SSH key files by file name, <name>.key and
	// <name>.pub, which the archive keeps under .ssh/.
	Keys map[string][]byte
	// WorkFactor is the scrypt work factor, log2 of N, Write uses;
	// zero is age's default.
	WorkFactor int
}

// Write encrypts the bundle to a passphrase into w.
func (b *Bundle) Write(w io.Writer, passphrase string) error {
	if err := CheckPassphrase(passphrase); err != nil {
		return err
	}
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return err
	}
	if b.WorkFactor > 0 {
		recipient.SetWorkFactor(b.WorkFactor)
	}
	encrypted, err := age.Encrypt(w, recipient)
	if err != nil {
		return err
	}
	zipped := gzip.NewWriter(encrypted)
	archive := tar.NewWriter(zipped)
	manifest, err := json.MarshalIndent(b.Manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := entry(archive, manifestName, manifest); err != nil {
		return err
	}
	names := make([]string, 0, len(b.Keys))
	for name := range b.Keys {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := entry(archive, sshDir+name, b.Keys[name]); err != nil {
			return err
		}
	}
	for _, closer := range []io.Closer{archive, zipped, encrypted} {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// entry writes one private file into the archive.
func entry(archive *tar.Writer, name string, data []byte) error {
	header := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(data)), ModTime: time.Now(), Typeflag: tar.TypeReg}
	if err := archive.WriteHeader(header); err != nil {
		return err
	}
	_, err := archive.Write(data)
	return err
}

// ReadBundle decrypts a bundle with its passphrase and reads it. It
// refuses a file over MaxBundle, an entry over its limit, an entry that
// is not a plain file, and a path other than manifest.json or a key file
// of .ssh/.
func ReadBundle(r io.Reader, passphrase string) (*Bundle, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, ErrPassphrase
	}
	limited := &io.LimitedReader{R: r, N: MaxBundle + 1}
	decrypted, err := age.Decrypt(limited, identity)
	var noMatch *age.NoIdentityMatchError
	if errors.As(err, &noMatch) {
		return nil, ErrPassphrase
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBundle, err)
	}
	zipped, err := gzip.NewReader(decrypted)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBundle, err)
	}
	bundle, err := readArchive(tar.NewReader(zipped))
	if limited.N <= 0 {
		return nil, fmt.Errorf("%w: over %d MiB", ErrBundle, MaxBundle>>20)
	}
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

// readArchive reads the manifest and the key files of an archive.
func readArchive(archive *tar.Reader) (*Bundle, error) {
	bundle := &Bundle{Keys: map[string][]byte{}}
	var manifest []byte
	for count := 0; ; count++ {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrBundle, err)
		}
		if count >= MaxEntries {
			return nil, fmt.Errorf("%w: over %d entries", ErrBundle, MaxEntries)
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		name, limit, err := entryName(header)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(archive, limit+1))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrBundle, err)
		}
		if int64(len(data)) > limit {
			return nil, fmt.Errorf("%w: %s is over %d bytes", ErrBundle, header.Name, limit)
		}
		if name == manifestName {
			manifest = data
			continue
		}
		bundle.Keys[name] = data
	}
	if manifest == nil {
		return nil, fmt.Errorf("%w: no %s", ErrBundle, manifestName)
	}
	if err := json.Unmarshal(manifest, &bundle.Manifest); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrBundle, manifestName, err)
	}
	if bundle.Manifest.Format != Format {
		return nil, fmt.Errorf("%w: format %d, this task-ui reads %d", ErrBundle, bundle.Manifest.Format, Format)
	}
	return bundle, nil
}

// entryName checks an entry and returns its name, a key file's relative
// to .ssh/, and the limit of its size.
func entryName(header *tar.Header) (string, int64, error) {
	name := header.Name
	if header.Typeflag != tar.TypeReg {
		return "", 0, fmt.Errorf("%w: %s is not a plain file", ErrBundle, name)
	}
	if name == manifestName {
		return name, MaxManifest, nil
	}
	base, ok := strings.CutPrefix(name, sshDir)
	stem := strings.TrimSuffix(strings.TrimSuffix(base, ".key"), ".pub")
	if !ok || path.Clean(name) != name || strings.Contains(base, "/") || stem == base || stem == "" || strings.HasPrefix(base, ".") {
		return "", 0, fmt.Errorf("%w: %s is neither manifest.json nor a .key or .pub file of .ssh", ErrBundle, name)
	}
	return base, MaxKeyFile, nil
}
