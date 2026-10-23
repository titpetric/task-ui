package secret

// PEMKeys are private keys held in memory, which decrypt and never
// encrypt: the encryption keys an older workspace bundle carries, which
// open the secrets it carries encrypted.
type PEMKeys [][]byte

// Recipient returns none: PEMKeys only decrypt.
func (p PEMKeys) Recipient() (string, string, bool) {
	return "", "", false
}

// Identities returns the keys.
func (p PEMKeys) Identities() [][]byte {
	return p
}
