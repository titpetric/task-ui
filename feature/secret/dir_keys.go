package secret

import (
	"os"
	"path/filepath"

	"github.com/titpetric/task-ui/sshkey"
)

// DirKeys are the keys of a folder of id_* files, as the instance kept
// them before the key folder: the legacy vault of a migration reads
// with them. A type's own key encrypts, id_ed25519 before id_rsa,
// never a backup or a key made for someone or a use, such as
// id_rsa_deploy; every private key in the folder decrypts.
type DirKeys struct {
	Store *sshkey.Store
}

// Recipient returns the folder's key that encrypts.
func (d DirKeys) Recipient() (string, string, bool) {
	for _, kind := range []string{"ed25519", "rsa"} {
		for _, key := range d.Store.List() {
			if key.Backup || key.Label != "" || sshkey.KindOf(key.Name) != kind {
				continue
			}
			if _, err := recipient(key.PublicKey); err == nil {
				return key.Name, key.PublicKey, true
			}
		}
	}
	return "", "", false
}

// Identities returns every private key of the folder.
func (d DirKeys) Identities() [][]byte {
	var result [][]byte
	for _, key := range d.Store.List() {
		if data, err := os.ReadFile(filepath.Join(d.Store.Dir, key.Name)); err == nil {
			result = append(result, data)
		}
	}
	return result
}
