// Package sshkey manages the key pairs in the instance SSH folder: the
// identities jobs and git use to reach other hosts.
package sshkey

import (
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
	// Backup is set for keys moved aside by a rotation.
	Backup bool `json:"backup,omitempty"`
}

// KindOf returns the key type a file name stands for, or "" for a name
// that is not one of the generated ones.
func KindOf(name string) string {
	for kind, file := range Types {
		if file == name {
			return kind
		}
	}
	return ""
}
