// Package testkey hands a test key material without generating any.
//
// sshkey.Store.Generate makes a 4096 bit RSA key, which takes 1.5
// seconds of CPU, measured, and nine tests asked for one. ed25519 and
// ecdsa are free and need none of this; RSA is the whole reason the
// package exists.
//
// A key is generated once into testdata and read from there ever after,
// by this run and by every run on this machine. The directory is
// gitignored: a private key in a repository reads as a leak even when it
// is a fixture, and the first run that needs one pays for it in a second
// and a half nobody pays again.
package testkey

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Bits is what sshkey.Store.Generate makes for RSA, so a cached key
// stands in for a generated one without changing the name a store gives
// it: deploy_rsa4096 either way.
const Bits = 4096

// cached keeps what this process already read, so a package of twenty
// tests reads the file once.
var cached sync.Map

// RSA is an RSA private key in PEM, of the size sshkey generates. Every
// caller gets the same key: a test that wants two different ones asks
// for a second size.
func RSA(t *testing.T) []byte { return Of(t, Bits) }

// Of is an RSA private key of a size, in PEM. 2048 is what a test that
// asserts on the name rsa2048 wants.
func Of(t *testing.T, bits int) []byte {
	t.Helper()
	if data, ok := cached.Load(bits); ok {
		return data.([]byte)
	}
	data, err := read(bits)
	if err != nil {
		t.Fatalf("the cached %d bit key: %v", bits, err)
	}
	cached.Store(bits, data)
	return data
}

// read answers the file in testdata, making it when it is not there.
//
// The write goes to a temporary name and is renamed over, so two test
// binaries racing for the first key leave one whole file rather than
// half of two.
func read(bits int) ([]byte, error) {
	path := filepath.Join(dir(), "rsa"+itoa(bits)+".key")
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	block, err := ssh.MarshalPrivateKey(key, "task-ui test key")
	if err != nil {
		return nil, err
	}
	data := pem.EncodeToMemory(block)
	if err := os.MkdirAll(dir(), 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(dir(), "rsa*.tmp")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return nil, err
	}
	return data, nil
}

// dir is this package's testdata, found from this file: a test of
// another package runs in its own directory and has no other way to
// name it.
func dir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata")
}

// itoa keeps the import list to what the keys need.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}
