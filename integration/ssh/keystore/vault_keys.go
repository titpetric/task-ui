package keystore

import (
	"context"
)

// VaultKeys are the encryption key of a scope as a vault takes it:
// secret.Keys over the key folder.
type VaultKeys struct {
	Store *Local
	Scope string
}

// Recipient returns the scope's encryption key.
func (v VaultKeys) Recipient() (string, string, bool) {
	key := v.Store.EncryptKey(context.Background(), v.Scope)
	if key == nil || key.PublicKey == "" {
		return "", "", false
	}
	return key.Name, key.PublicKey, true
}

// Identities returns the private halves of the scope's encryption key,
// its backups after.
func (v VaultKeys) Identities() [][]byte {
	key := v.Store.EncryptKey(context.Background(), v.Scope)
	if key == nil {
		return nil
	}
	return v.Store.identities(key.Name)
}
