//go:build integration

package sshkey

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_Upload(t *testing.T) {
	s := &Store{Dir: t.TempDir()}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)})
	key, err := s.Upload([]byte(strings.ReplaceAll(string(pemData), "\n", "\r\n")), "laptop")
	if err != nil {
		t.Fatal(err)
	}
	if key.Name != "id_rsa" || key.Bits != 2048 || key.Comment != "laptop" || key.Type != "ssh-rsa" {
		t.Fatalf("uploaded: %+v", key)
	}

	// A key generated elsewhere uploads, in OpenSSH format.
	other := &Store{Dir: t.TempDir()}
	if _, err := other.Generate("ed25519", "elsewhere"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(other.Dir, "id_ed25519"))
	uploaded, err := s.Upload(data, "")
	if err != nil || uploaded.Name != "id_ed25519" {
		t.Fatalf("ed25519 upload: %+v %v", uploaded, err)
	}

	// Upload refuses a key behind a passphrase: runs have nobody to type it.
	if _, err := exec.LookPath("ssh-keygen"); err == nil {
		locked := filepath.Join(t.TempDir(), "locked")
		cmd := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "secret", "-f", locked)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("ssh-keygen: %v\n%s", err, out)
		}
		data, _ := os.ReadFile(locked)
		if _, err := s.Upload(data, ""); !errors.Is(err, ErrPassphrase) {
			t.Fatalf("passphrase key: %v", err)
		}
	}
	if _, err := s.Upload([]byte("not a key"), ""); err == nil {
		t.Fatal("accepted garbage")
	}
}
