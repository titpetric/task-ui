package sshagent

// Named is a private key by name, in PEM or OpenSSH form, held in memory.
type Named struct {
	Name string
	PEM  []byte
}
