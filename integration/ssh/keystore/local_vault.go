package keystore

// Vault returns the encryption key of a scope for a vault.
func (l *Local) Vault(scope string) VaultKeys {
	return VaultKeys{Store: l, Scope: scope}
}
