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

// Extension ends the file name of a bundle: <folder>.taskui-ws.
const Extension = ".taskui-ws"

// The limits ReadBundle holds a bundle to: the encrypted file, the manifest,
// each key file and the number of entries.
const (
	MaxBundle   = 64 << 20
	MaxManifest = 16 << 20
	MaxKeyFile  = 1 << 20
	MaxEntries  = 1024
)

// manifestName and keysDir are where a bundle keeps the manifest and
// the key files.
const (
	manifestName = "manifest.json"
	keysDir      = "keys/"
)

// ErrPassphrase reports a passphrase that does not open a bundle.
var ErrPassphrase = errors.New("the passphrase does not open the bundle")

// ErrBundle reports a file that is not a bundle, or one that breaks its
// limits.
var ErrBundle = errors.New("not a workspace bundle")

// Bundle is a workspace as one instance exports it for another: a tar.gz
// of manifest.json and keys/, the workspace's key files, encrypted to
// a passphrase with age's scrypt.
type Bundle struct {
	Manifest Manifest
	// Keys are the key files by their path in the keyring's directory
	// of the workspace: encrypt/<name> and ssh/<name>.
	Keys map[string][]byte
	// WorkFactor is the scrypt work factor, log2 of N, Write uses;
	// zero is age's default.
	WorkFactor int
}

// Write encrypts the bundle to a passphrase into w.
func (b *Bundle) Write(w io.Writer, passphrase string) error {
	if passphrase == "" {
		return errors.New("a bundle needs a passphrase")
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
		if err := entry(archive, keysDir+name, b.Keys[name]); err != nil {
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
// is not a plain file, and a path that would leave keys/encrypt or
// keys/ssh.
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
// to keys/, and the limit of its size.
func entryName(header *tar.Header) (string, int64, error) {
	name := header.Name
	if header.Typeflag != tar.TypeReg {
		return "", 0, fmt.Errorf("%w: %s is not a plain file", ErrBundle, name)
	}
	if name == manifestName {
		return name, MaxManifest, nil
	}
	rel, ok := strings.CutPrefix(name, keysDir)
	dir, base := path.Split(rel)
	if !ok || path.Clean(name) != name || (dir != "encrypt/" && dir != "ssh/") || base == "" || strings.HasPrefix(base, ".") {
		return "", 0, fmt.Errorf("%w: %s is outside manifest.json, keys/encrypt and keys/ssh", ErrBundle, name)
	}
	return rel, MaxKeyFile, nil
}
