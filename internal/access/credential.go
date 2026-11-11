package access

// Credential is one line of the password file: a name and the bcrypt
// hash of the password.
type Credential struct {
	Name string
	Hash string
}

// Line writes the credential the way the password file holds it.
func (c Credential) Line() string {
	return c.Name + ":" + c.Hash
}
