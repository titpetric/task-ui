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

// A key made for someone is its type with a label: KindOf reads the kind and
// label back from the name, and a rotation keeps them.
func TestKindOf_Labelled(t *testing.T) {
	if KindOf("id_ed25519_alice") != "ed25519" || labelOf("id_ed25519_alice") != "alice" {
		t.Fatal("kind and label of a labelled key")
	}
	if KindOf("id_rsa_ops-team") != "rsa" || labelOf("id_rsa_ops-team") != "ops-team" || labelOf("id_rsa") != "" {
		t.Fatal("labels")
	}
	if KindOf("id_ed25519_alice.bak-20260101T000000") != "" {
		t.Fatal("a backup has no kind")
	}
}
