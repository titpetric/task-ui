package sshkey

import (
	"testing"
)

func TestKindOf(t *testing.T) {
	if KindOf("id_ed25519") != "ed25519" || KindOf("id_rsa") != "rsa" || KindOf("id_ecdsa") != "ecdsa" {
		t.Fatal("kind of generated names")
	}
	if KindOf("id_ed25519.bak-20260101T000000") != "" || KindOf("other") != "" {
		t.Fatal("kind of other names")
	}
	key := Key{Name: "id_ed25519", Type: "ssh-ed25519"}
	if key.Backup {
		t.Fatal("backup by default")
	}
}
