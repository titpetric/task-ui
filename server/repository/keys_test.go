package repository

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestKeysGenerateRotateDelete(t *testing.T) {
	keys := &Keys{Dir: filepath.Join(t.TempDir(), ".ssh")}

	first, err := keys.Generate("ed25519", "task ui")
	if err != nil {
		t.Fatal(err)
	}
	if first.Name != "id_ed25519" || first.Type != "ssh-ed25519" || first.Comment != "task ui" {
		t.Fatalf("generated: %+v", first)
	}
	info, err := os.Stat(filepath.Join(keys.Dir, "id_ed25519"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode: %v %v", info.Mode(), err)
	}
	data, _ := os.ReadFile(filepath.Join(keys.Dir, "id_ed25519"))
	if _, err := ssh.ParsePrivateKey(data); err != nil {
		t.Fatalf("written key does not parse: %v", err)
	}

	second, err := keys.Generate("ed25519", "rotated")
	if err != nil {
		t.Fatal(err)
	}
	if second.Fingerprint == first.Fingerprint {
		t.Fatal("rotation kept the same key")
	}

	list := keys.List()
	if len(list) != 2 || list[0].Backup || !list[1].Backup || list[1].Fingerprint != first.Fingerprint {
		t.Fatalf("list after rotation: %+v", list)
	}

	if err := keys.Delete(list[1].Name); err != nil {
		t.Fatal(err)
	}
	if err := keys.Delete("../../etc/passwd"); err == nil {
		t.Fatal("deleted a path outside the key folder")
	}
	if len(keys.List()) != 1 {
		t.Fatal("backup not deleted")
	}

	if _, err := keys.Generate("dsa", ""); err == nil {
		t.Fatal("generated an unsupported key type")
	}
}

func TestKeysUpload(t *testing.T) {
	keys := &Keys{Dir: t.TempDir()}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)})
	key, err := keys.Upload([]byte(strings.ReplaceAll(string(pemData), "\n", "\r\n")), "laptop")
	if err != nil {
		t.Fatal(err)
	}
	if key.Name != "id_rsa" || key.Bits != 2048 || key.Comment != "laptop" {
		t.Fatalf("uploaded: %+v", key)
	}

	other := &Keys{Dir: t.TempDir()}
	block, _ := ssh.MarshalPrivateKey(mustEd25519(t), "")
	uploaded, err := other.Upload(pem.EncodeToMemory(block), "")
	if err != nil || uploaded.Name != "id_ed25519" {
		t.Fatalf("ed25519 upload: %+v %v", uploaded, err)
	}

	encrypted, err := ssh.MarshalPrivateKeyWithPassphrase(private, "", []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Upload(pem.EncodeToMemory(encrypted), ""); !errors.Is(err, ErrPassphrase) {
		t.Fatalf("passphrase key: %v", err)
	}
	if _, err := keys.Upload([]byte("not a key"), ""); err == nil {
		t.Fatal("accepted garbage")
	}
}
