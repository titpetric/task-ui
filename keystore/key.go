// Package keystore keeps every key of the instance in one folder, flat:
// <name>.key, the private key, and <name>.pub, beside known_hosts. The
// name says the algorithm and the size, as deploy_rsa4096 or
// task_encryption_ed25519, and the database says what a key is: an
// encryption key or an SSH key, of the instance or a workspace, and
// whose runs get an SSH key. Nothing else reads a key file: git, the
// inventory's logins, the secrets and the runs go through the store,
// and a run gets its SSH keys in an ssh agent of its own, in memory.
package keystore

import (
	"strings"
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

// ShortPublic is the public key as a table shows it: its type and the
// end of its material, where keys of a type differ, the comment left
// out. PublicKey is
// the whole line, to copy.
func (k Key) ShortPublic() string {
	fields := strings.Fields(k.PublicKey)
	if len(fields) < 2 {
		return k.PublicKey
	}
	material := fields[1]
	if len(material) > 10 {
		material = "…" + material[len(material)-10:]
	}
	return fields[0] + " " + material
}
