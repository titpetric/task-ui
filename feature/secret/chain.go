package secret

// Chain is a list of key sets as one: the first's recipient encrypts,
// and every set's identities decrypt, in order. A migration reads with
// the SSH keys of old and the workspaces' encryption keys at once.
type Chain []Keys

// Recipient returns the first set's key that encrypts.
func (c Chain) Recipient() (string, string, bool) {
	for _, keys := range c {
		if name, public, ok := keys.Recipient(); ok {
			return name, public, true
		}
	}
	return "", "", false
}

// Identities returns every set's private keys, in order.
func (c Chain) Identities() [][]byte {
	var result [][]byte
	for _, keys := range c {
		result = append(result, keys.Identities()...)
	}
	return result
}
