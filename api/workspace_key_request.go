package api

// WorkspaceKeyRequest generates a workspace key, or uploads a private
// key, or sets whose runs get one of the workspace's SSH keys.
type WorkspaceKeyRequest struct {
	// Encrypt points the request at the workspace's encryption key:
	// generate makes it, or rotates it to new material; upload takes one when
	// there is none. The previous material stays as a backup.
	Encrypt bool `json:"encrypt"`
	// Algo, or Type, and Bits are the SSH key to generate: ed25519, the
	// default, ecdsa or rsa.
	Algo string `json:"algo"`
	Type string `json:"type"`
	Bits int    `json:"bits"`
	// Name is the key's name before workspace_<folder>_ and its
	// algorithm: build names workspace_malta_build_ed25519. Empty is ssh.
	Name    string `json:"name"`
	Comment string `json:"comment"`
	// Content is the private key to upload, PEM or OpenSSH.
	Content string `json:"content"`
	// Policies are the roles, and Users the users by name, whose runs
	// in the workspace get an SSH key. With neither, every run in the
	// workspace gets it.
	Policies []string `json:"policies"`
	Users    []string `json:"users"`
}

// keyRequest returns the request as a request of the key folder.
func (req *WorkspaceKeyRequest) keyRequest() KeyRequest {
	return KeyRequest{Name: req.Name, Algo: req.Algo, Type: req.Type, Bits: req.Bits, Comment: req.Comment, Content: req.Content, Policies: req.Policies, Users: req.Users}
}
