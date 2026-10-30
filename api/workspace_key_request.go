package api

import (
	"errors"
)

// ErrWorkspaceEncrypt refuses an encryption key for a workspace: the
// instance's one encryption key encrypts every secret.
var ErrWorkspaceEncrypt = errors.New("a workspace has no encryption key: the instance's encryption key encrypts every secret, one per instance")

// WorkspaceKeyRequest generates a workspace key, or uploads a private
// key as one.
//
// Whose runs get it is no field. The key is added to one workspace and
// every run of that workspace gets it: the scope is the grant, and
// there is nothing left to narrow.
type WorkspaceKeyRequest struct {
	// Encrypt names a workspace's encryption key, which no workspace
	// has any more: true answers 400, ErrWorkspaceEncrypt.
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
}

// keyRequest returns the request as a request of the key folder, with
// no links: a workspace's key is handed out by its scope.
func (req *WorkspaceKeyRequest) keyRequest() KeyRequest {
	return KeyRequest{Name: req.Name, Algo: req.Algo, Type: req.Type, Bits: req.Bits, Comment: req.Comment, Content: req.Content}
}
