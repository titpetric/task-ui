package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// KeyRequest generates a key pair, or uploads a private key.
type KeyRequest struct {
	// Type is the pair to generate: ed25519, ecdsa or rsa.
	Type string `json:"type"`
	// Name labels a key made for someone: id_<type>_<name>, beside the
	// type's own key. Empty is the type's own key.
	Name    string `json:"name"`
	Comment string `json:"comment"`
	// Content is the private key to upload, PEM encoded.
	Content string `json:"content"`
	// Policies are the roles whose runs get the key through the run's
	// ssh agent. Left out, a key keeps the roles it had, so a rotation
	// keeps them; an empty list withholds the key from all.
	Policies []string `json:"policies"`
	// Users are the users, by name, whose runs get the key, beside the
	// roles. Left out, a key keeps the users it had.
	Users []string `json:"users"`
}

// listKeys returns the SSH key pairs of the instance SSH folder, the
// encryption keys left out, beside the instance's encryption key and
// the keys held for both uses. Takes key/view.
func (h *Handlers) listKeys(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/view") {
		return
	}
	policies, err := h.opts.KeyPolicies.All(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	users, err := h.opts.KeyPolicies.AllUsers(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	body := map[string]any{"keys": h.accessKeys(), "dir": h.opts.Keys.Dir, "policies": policies, "users": users}
	if keyring := h.opts.WorkspaceKeys; keyring != nil {
		body["encrypt"], body["overlaps"] = keyring.InstanceKey(), keyring.Overlaps()
	}
	writeJSON(w, http.StatusOK, body)
}

// accessKeys returns the key pairs of the instance SSH folder, backups
// among them, without a key that encrypts: a copy of an encryption key
// is no SSH key, whatever folder holds it.
func (h *Handlers) accessKeys() []sshkey.Key {
	keys := h.opts.Keys.List()
	if h.opts.WorkspaceKeys == nil {
		return keys
	}
	result := []sshkey.Key{}
	for _, key := range keys {
		if !h.opts.WorkspaceKeys.Encrypts(filepath.Join(h.opts.Keys.Dir, key.Name)) {
			result = append(result, key)
		}
	}
	return result
}

// checkAccess refuses an upload that is an encryption key, ErrRole.
func (h *Handlers) checkAccess(content string) error {
	if h.opts.WorkspaceKeys == nil {
		return nil
	}
	return h.opts.WorkspaceKeys.CheckAccess([]byte(content))
}

// setPolicies sets the roles and the users whose runs get a key, each
// when the request names them.
func (h *Handlers) setPolicies(r *http.Request, name string, req KeyRequest) error {
	if req.Policies != nil {
		if err := h.opts.KeyPolicies.Set(r.Context(), name, req.Policies); err != nil {
			return err
		}
	}
	if req.Users != nil {
		users := make([]string, 0, len(req.Users))
		for _, user := range req.Users {
			users = append(users, strings.TrimSpace(user))
		}
		return h.opts.KeyPolicies.SetUsers(r.Context(), name, users)
	}
	return nil
}

// updateKey changes the roles whose runs get a key. Takes key/edit.
func (h *Handlers) updateKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") {
		return
	}
	name := param(r, "name")
	found := false
	for _, key := range h.accessKeys() {
		found = found || (key.Name == name && !key.Backup)
	}
	if !found {
		failErr(w, storage.ErrNotFound)
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Policies == nil {
		req.Policies = []string{}
	}
	if req.Users == nil {
		req.Users = []string{}
	}
	if err := h.setPolicies(r, name, req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": name, "policies": req.Policies, "users": req.Users})
}

// generateKey makes a new pair. It keeps the existing pair of the same
// type as a backup. An SSH key encrypts nothing, so no secret changes.
func (h *Handlers) generateKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") {
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	key, err := h.opts.Keys.GenerateAs(req.Type, strings.TrimSpace(req.Name), strings.TrimSpace(req.Comment))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.setPolicies(r, key.Name, req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key})
}

// uploadKey stores a private key as the pair of its type. A key the
// instance or a workspace encrypts with answers 409.
func (h *Handlers) uploadKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") {
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.checkAccess(req.Content); err != nil {
		failErr(w, err)
		return
	}
	key, err := h.opts.Keys.UploadAs([]byte(req.Content), strings.TrimSpace(req.Name), strings.TrimSpace(req.Comment))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.setPolicies(r, key.Name, req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key})
}

// deleteKey removes a pair, or a backup, and its policies.
func (h *Handlers) deleteKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/delete") {
		return
	}
	name := param(r, "name")
	if err := h.opts.Keys.Delete(name); err != nil {
		failErr(w, err)
		return
	}
	if err := h.opts.KeyPolicies.Delete(r.Context(), name); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// rotateEncryptKey generates a new encryption key for the instance and
// writes the services' secrets, the integrations' settings, again for
// it. The previous key stays as a backup and still opens them, so a
// rekey that fails sets rekey_error. Takes key/edit.
func (h *Handlers) rotateEncryptKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/edit") {
		return
	}
	if h.opts.WorkspaceKeys == nil {
		fail(w, http.StatusNotFound, "the keyring is off")
		return
	}
	comment := "task-ui instance"
	if key := h.opts.WorkspaceKeys.InstanceKey(); key != nil && key.Comment != "" {
		comment = key.Comment
	}
	key, err := h.opts.WorkspaceKeys.RotateInstance(comment)
	if err != nil {
		failErr(w, err)
		return
	}
	n, err := h.opts.Secrets.RekeyInstance(r.Context())
	body := map[string]any{"key": key, "rekeyed": n}
	if err != nil {
		body["rekey_error"] = err.Error()
	}
	writeJSON(w, http.StatusCreated, body)
}
