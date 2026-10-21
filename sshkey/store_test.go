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

func TestStore_BackupSameSecond(t *testing.T) {
	t.Parallel()
	s := &Store{Dir: t.TempDir()}
	seen := map[string]bool{}
	for i := 0; i < 4; i++ {
		key, err := s.Generate("ed25519", "")
		if err != nil {
			t.Fatal(err)
		}
		seen[key.Fingerprint] = true
	}
	list := s.List()
	if len(list) != 4 || len(seen) != 4 {
		t.Fatalf("four rotations in a second keep every key: %+v", list)
	}
	for _, key := range list {
		if !seen[key.Fingerprint] {
			t.Fatalf("unknown key %+v", key)
		}
		if key.Backup && (!strings.Contains(key.PublicKey, "ssh-ed25519") || KindOf(key.Name) != "") {
			t.Fatalf("backup %+v", key)
		}
	}
	if err := s.Delete(list[3].Name); err != nil {
		t.Fatal(err)
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

// A key made for someone sits beside the type's own key: generating or
// uploading one under a label leaves the other alone, and rotating it
// keeps its name.
func TestStore_GenerateAs(t *testing.T) {
	k := &Store{Dir: t.TempDir()}
	own, err := k.Generate("ed25519", "instance")
	if err != nil {
		t.Fatal(err)
	}
	alice, err := k.GenerateAs("ed25519", "alice", "alice@task")
	if err != nil {
		t.Fatal(err)
	}
	if alice.Name != "id_ed25519_alice" || alice.Kind != "ed25519" || alice.Label != "alice" {
		t.Fatalf("labelled key: %+v", alice)
	}
	if own.Label != "" || own.Kind != "ed25519" {
		t.Fatalf("own key: %+v", own)
	}
	again, err := k.GenerateAs("ed25519", "alice", "alice@task")
	if err != nil || again.Fingerprint == alice.Fingerprint {
		t.Fatal("rotating the labelled key makes a new one")
	}
	names := map[string]bool{}
	for _, key := range k.List() {
		names[key.Name] = true
	}
	if !names["id_ed25519"] || !names["id_ed25519_alice"] || len(names) != 3 {
		t.Fatalf("own key, labelled key and its backup: %v", names)
	}
	for _, bad := range []string{"Alice", "a b", "../x", "-x"} {
		if _, err := k.GenerateAs("ed25519", bad, ""); err == nil {
			t.Fatalf("label %q accepted", bad)
		}
	}
	if err := k.Delete("id_ed25519_alice"); err != nil {
		t.Fatal(err)
	}
}

func TestStore_UploadAs(t *testing.T) {
	k := &Store{Dir: t.TempDir()}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	data := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	got, err := k.UploadAs(data, "ops", "ops key")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "id_rsa_ops" || got.Label != "ops" {
		t.Fatalf("uploaded under a label: %+v", got)
	}
	if _, err := k.UploadAs(data, "Not OK", ""); err == nil {
		t.Fatal("a bad label is refused")
	}
}
