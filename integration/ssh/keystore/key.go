package keystore

import (
	"time"
)

// Key is a key of the folder, as the pages and the API show it: never
// its private half.
type Key struct {
	// Name is the key's name and the stem of its files.
	Name string `json:"name"`
	// Purpose is encrypt or ssh; Scope instance or workspace:<id>.
	Purpose string `json:"purpose"`
	Scope   string `json:"scope"`
	// Algo is ed25519, ecdsa or rsa, and Bits its size, 0 for ed25519.
	Algo        string `json:"algo"`
	Bits        int    `json:"bits,omitempty"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
	// PublicKey is the authorized_keys line.
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
	// Roles and Users get an SSH key in their runs.
	Roles []string `json:"roles"`
	Users []string `json:"users"`
	// Git marks the SSH key git pulls and clones with.
	Git bool `json:"git,omitempty"`
	// Backup names the backup of the previous material a rotation kept,
	// backup/<name>, "" for none, and BackupFingerprint its fingerprint.
	// Until an admin deletes it, the key serves both and rotates no more.
	Backup            string     `json:"backup,omitempty"`
	BackupFingerprint string     `json:"backup_fingerprint,omitempty"`
	RotatedAt         *time.Time `json:"rotated_at,omitempty"`
}

// Workspace returns the ID of the workspace a key belongs to, "" for an
// instance key.
func (k Key) Workspace() string {
	if len(k.Scope) > len(scopeWorkspace) && k.Scope[:len(scopeWorkspace)] == scopeWorkspace {
		return k.Scope[len(scopeWorkspace):]
	}
	return ""
}

// Encrypts reports an encryption key.
func (k Key) Encrypts() bool {
	return k.Purpose == PurposeEncrypt
}

// Kind returns the algorithm and the size the name ends with: ed25519, ecdsa256, rsa4096.
func (k Key) Kind() string {
	return tag(k.Algo, k.Bits)
}

// Size is the key's size in bits: Bits, 256 for ed25519.
func (k Key) Size() int {
	if k.Bits == 0 && k.Algo == "ed25519" {
		return 256
	}
	return k.Bits
}
