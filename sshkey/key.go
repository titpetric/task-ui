// Package sshkey manages the key pairs in the instance SSH folder: the
// identities jobs and git use to reach other hosts.
package sshkey

import (
	"strings"
	"time"
)

// Key describes a key pair in the instance SSH folder.
type Key struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Bits        int       `json:"bits,omitempty"`
	Fingerprint string    `json:"fingerprint"`
	Comment     string    `json:"comment"`
	PublicKey   string    `json:"public_key"`
	Modified    time.Time `json:"modified"`
	// Backup marks keys a rotation moved aside.
	Backup bool `json:"backup,omitempty"`
	// Kind is the key type the name stands for, ed25519, ecdsa or rsa,
	// and Label what follows it, for a key made for someone:
	// id_ed25519_alice is ed25519 and alice. A rotation keeps both.
	Kind  string `json:"kind,omitempty"`
	Label string `json:"label,omitempty"`
}

// KindOf returns the key type a file name stands for, its own key or one
// under a label; "" for a backup or a name that is not one of them.
func KindOf(name string) string {
	if strings.Contains(name, ".bak-") {
		return ""
	}
	for kind, file := range Types {
		if name == file || strings.HasPrefix(name, file+"_") {
			return kind
		}
	}
	return ""
}

// labelOf returns what follows the type in a key's name, "" for the
// type's own key.
func labelOf(name string) string {
	if kind := KindOf(name); kind != "" {
		return strings.TrimPrefix(strings.TrimPrefix(name, Types[kind]), "_")
	}
	return ""
}
