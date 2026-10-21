package api

// WorkspaceKeyRequest generates a workspace key pair, or uploads a
// private key, or sets whose runs get one of the workspace's SSH keys.
type WorkspaceKeyRequest struct {
	// Encrypt points the request at the workspace's encryption key,
	// always ed25519: generate rotates it, upload replaces it; the
	// previous one stays as a backup.
	Encrypt bool `json:"encrypt"`
	// Type is the SSH key pair to generate: ed25519, ecdsa or rsa.
	Type string `json:"type"`
	// Name labels an SSH key: id_<type>_<name>. Empty is the type's
	// own key.
	Name    string `json:"name"`
	Comment string `json:"comment"`
	// Content is the private key to upload, PEM encoded.
	Content string `json:"content"`
	// Policies are the roles, and Users the users by name, whose runs
	// in the workspace get an SSH key. With neither, every run in the
	// workspace gets it. Left out on generate or upload, a key keeps
	// the grants it had, so a rotation keeps them.
	Policies []string `json:"policies"`
	Users    []string `json:"users"`
}
