package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/storage"
)

// KeyRequest generates a key, uploads a private key, or sets whose runs
// get an SSH key and whether git pulls with it.
type KeyRequest struct {
	// Name is the key's name before its algorithm: deploy names
	// deploy_rsa4096. Unique across the key folder.
	Name string `json:"name"`
	// Algo and Bits are what to generate: ed25519, the default, ecdsa of
	// 256, 384 or 521 bits, rsa of 2048, 3072 or 4096. Type is Algo too.
	Algo string `json:"algo"`
	Type string `json:"type"`
	Bits int    `json:"bits"`
	// Purpose is ssh, the default, or encrypt.
	Purpose string `json:"purpose"`
	Comment string `json:"comment"`
	// Content is the private key to upload, PEM or OpenSSH.
	Content string `json:"content"`
	// Policies are the roles, and Users the users by name, whose runs
	// get an SSH key through the run's ssh agent. Left out on a PUT, a
	// key keeps them.
	Policies []string `json:"policies"`
	Users    []string `json:"users"`
	// Git makes an SSH key the one git pulls and clones with.
	Git *bool `json:"git,omitempty"`
}

// spec returns what the request makes, in a scope.
func (req *KeyRequest) spec(scope string) keystore.Spec {
	algo := req.Algo
	if algo == "" {
		algo = req.Type
	}
	return keystore.Spec{Name: strings.TrimSpace(req.Name), Algo: algo, Bits: req.Bits, Purpose: req.Purpose, Scope: scope, Comment: strings.TrimSpace(req.Comment)}
}

// keyStatus maps an error of the key folder to its status.
func keyStatus(err error) int {
	switch {
	case errors.Is(err, keystore.ErrExists):
		return http.StatusConflict
	case errors.Is(err, storage.ErrNotFound):
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// failKey answers an error of the key folder.
func failKey(w http.ResponseWriter, err error) {
	fail(w, keyStatus(err), err.Error())
}

// keysOff answers a request on an instance without the key folder.
func (h *Handlers) keysOff(w http.ResponseWriter) bool {
	if h.opts.Keys == nil {
		fail(w, http.StatusNotFound, "the key folder isn't available on this instance")
		return true
	}
	return false
}

// listKeys returns every key of the key folder, public halves and
// fingerprints only: purpose, algorithm and size, scope, the roles and
// users an SSH key goes to, the one git pulls with, and the reads of each
// over MCP or the API, with the last read. Takes key/view.
func (h *Handlers) listKeys(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/view") || h.keysOff(w) {
		return
	}
	keys, err := h.opts.Keys.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	body := map[string]any{"keys": keys, "dir": h.opts.Keys.Dir}
	if h.opts.KeyReads != nil {
		counts, err := h.opts.KeyReads.Counts(r.Context())
		if err != nil {
			failErr(w, err)
			return
		}
		last, err := h.opts.KeyReads.Last(r.Context())
		if err != nil {
			failErr(w, err)
			return
		}
		body["reads"], body["last_read"] = counts, last
	}
	if git := h.opts.Keys.Git(r.Context()); git != nil {
		body["git"] = git.Name
	}
	if key := h.opts.Keys.EncryptKey(r.Context(), storage.ScopeInstance); key != nil {
		body["encrypt"] = key
	}
	writeJSON(w, http.StatusOK, body)
}

// link sets the roles and the users of a key and the git key, each
// when the request names it.
func (h *Handlers) link(r *http.Request, name string, req KeyRequest) error {
	if req.Policies != nil || req.Users != nil {
		key, err := h.opts.Keys.Get(r.Context(), name)
		if err != nil {
			return err
		}
		roles, users := key.Roles, key.Users
		if req.Policies != nil {
			roles = req.Policies
		}
		if req.Users != nil {
			users = trimmed(req.Users)
		}
		if err := h.opts.Keys.Link(r.Context(), name, roles, users); err != nil {
			return err
		}
	}
	if req.Git != nil {
		git := ""
		if *req.Git {
			git = name
		} else if current := h.opts.Keys.Git(r.Context()); current != nil && current.Name != name {
			git = current.Name
		}
		return h.opts.Keys.SetGit(r.Context(), git)
	}
	return nil
}

// trimmed returns the items trimmed.
func trimmed(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, strings.TrimSpace(item))
	}
	return result
}

// updateKey sets the roles and users whose runs get an SSH key, and
// whether git pulls with it. Takes key/edit.
func (h *Handlers) updateKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") || h.keysOff(w) {
		return
	}
	name := param(r, "name")
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.link(r, name, req); err != nil {
		failKey(w, err)
		return
	}
	key, err := h.opts.Keys.Get(r.Context(), name)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"key": key, "name": name, "policies": key.Roles, "users": key.Users})
}

// generateKey makes an instance key under a name of its own. Takes
// key/edit.
func (h *Handlers) generateKey(w http.ResponseWriter, r *http.Request) {
	h.addKey(w, r, func(req KeyRequest) (*keystore.Key, error) {
		return h.opts.Keys.Generate(r.Context(), req.spec(storage.ScopeInstance))
	})
}

// uploadKey stores a private key as an instance key; its algorithm and
// size come from the key. A key the folder holds, under any name or
// purpose, answers 409. Takes key/edit.
func (h *Handlers) uploadKey(w http.ResponseWriter, r *http.Request) {
	h.addKey(w, r, func(req KeyRequest) (*keystore.Key, error) {
		return h.opts.Keys.Import(r.Context(), req.spec(storage.ScopeInstance), []byte(req.Content))
	})
}

// addKey reads a key request, adds the key and links it.
func (h *Handlers) addKey(w http.ResponseWriter, r *http.Request, add func(KeyRequest) (*keystore.Key, error)) {
	if !h.require(w, r, "key/edit") || h.keysOff(w) {
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	key, err := add(req)
	if err != nil {
		failKey(w, err)
		return
	}
	if err := h.link(r, key.Name, req); err != nil {
		failKey(w, err)
		return
	}
	if key, err = h.opts.Keys.Get(r.Context(), key.Name); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key})
}

// deleteKey retires a key and drops its links. The encryption key of a
// scope that holds secrets stays: it opens them. Takes key/delete.
func (h *Handlers) deleteKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/delete") || h.keysOff(w) {
		return
	}
	name := param(r, "name")
	key, err := h.opts.Keys.Get(r.Context(), name)
	if err != nil {
		failErr(w, err)
		return
	}
	if key.Encrypts() {
		fail(w, http.StatusConflict, "an encryption key opens its secrets: rotate it or upload another in its place")
		return
	}
	if err := h.opts.Keys.Remove(r.Context(), name); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// rotateEncryptKey rotates the instance's encryption key to new material and
// writes the services' secrets, the integrations' settings, again for
// it. The previous material stays as a backup and still opens them, so
// a rekey that fails sets rekey_error. Takes key/edit.
func (h *Handlers) rotateEncryptKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") || h.keysOff(w) {
		return
	}
	current := h.opts.Keys.EncryptKey(r.Context(), storage.ScopeInstance)
	var key *keystore.Key
	var err error
	if current == nil {
		key, err = h.opts.Keys.Generate(r.Context(), keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt, Comment: "task-ui instance"})
	} else {
		key, err = h.opts.Keys.Rotate(r.Context(), current.Name)
	}
	if err != nil {
		failKey(w, err)
		return
	}
	n, err := h.opts.Secrets.RekeyInstance(r.Context())
	body := map[string]any{"key": key, "rekeyed": n}
	if err != nil {
		body["rekey_error"] = err.Error()
	}
	writeJSON(w, http.StatusCreated, body)
}
