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

func TestStore_Generate(t *testing.T) {
	s := &Store{Dir: filepath.Join(t.TempDir(), ".ssh")}

	first, err := s.Generate("ed25519", "task ui")
	if err != nil {
		t.Fatal(err)
	}
	if first.Name != "id_ed25519" || first.Type != "ssh-ed25519" || first.Comment != "task ui" || first.Bits != 256 {
		t.Fatalf("generated: %+v", first)
	}
	info, err := os.Stat(filepath.Join(s.Dir, "id_ed25519"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode: %v %v", info, err)
	}
	if !strings.HasPrefix(first.PublicKey, "ssh-ed25519 ") || !strings.HasSuffix(first.PublicKey, " task ui") {
		t.Fatalf("public key: %q", first.PublicKey)
	}

	second, err := s.Generate("ed25519", "rotated")
	if err != nil {
		t.Fatal(err)
	}
	if second.Fingerprint == first.Fingerprint {
		t.Fatal("rotation kept the same key")
	}
	list := s.List()
	if len(list) != 2 || list[0].Backup || !list[1].Backup || list[1].Fingerprint != first.Fingerprint {
		t.Fatalf("list after rotation: %+v", list)
	}

	if ecdsa, err := s.Generate("ecdsa", ""); err != nil || ecdsa.Type != "ecdsa-sha2-nistp256" || ecdsa.Bits != 256 {
		t.Fatalf("ecdsa: %+v %v", ecdsa, err)
	}
	if _, err := s.Generate("dsa", ""); err == nil {
		t.Fatal("generated an unsupported key type")
	}
}

func TestStore_Delete(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	if _, err := s.Generate("ed25519", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Generate("ed25519", "b"); err != nil {
		t.Fatal(err)
	}
	list := s.List()
	if err := s.Delete(list[1].Name); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("../../etc/passwd"); err == nil {
		t.Fatal("deleted a path outside the key folder")
	}
	if err := s.Delete("id_rsa"); err == nil {
		t.Fatal("deleted a missing key")
	}
	if len(s.List()) != 1 {
		t.Fatal("backup not deleted")
	}
}

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

	// A key behind a passphrase is refused: runs are unattended.
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
