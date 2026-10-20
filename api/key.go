package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// listKeys returns the key pairs of the instance SSH folder. Admins only.
func (h *Handlers) listKeys(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
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
	writeJSON(w, http.StatusOK, map[string]any{"keys": h.opts.Keys.List(), "dir": h.opts.Keys.Dir, "policies": policies, "users": users})
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

// updateKey changes the roles whose runs get a key. Admins only.
func (h *Handlers) updateKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	name := param(r, "name")
	found := false
	for _, key := range h.opts.Keys.List() {
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
// type as a backup.
func (h *Handlers) generateKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
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
	h.rekeyed(w, r, http.StatusCreated, key)
}

// rekeyed answers a key change with the key and the count of secrets
// written again for the keys now current. A rekey that fails is in the
// answer, not a failure: the key is there, the backups still open the
// secrets.
func (h *Handlers) rekeyed(w http.ResponseWriter, r *http.Request, status int, key any) {
	n, err := h.opts.Secrets.Rekey(r.Context())
	body := map[string]any{"key": key, "rekeyed": n}
	if err != nil {
		body["rekey_error"] = err.Error()
	}
	writeJSON(w, status, body)
}

// uploadKey stores a private key as the pair of its type.
func (h *Handlers) uploadKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
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
	h.rekeyed(w, r, http.StatusCreated, key)
}

// deleteKey removes a pair, or a backup. It first writes the secrets
// again for the keys that remain; the last key that can encrypt them
// stays until another is there.
func (h *Handlers) deleteKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	name := param(r, "name")
	if err := h.releaseKey(r.Context(), name); err != nil {
		fail(w, http.StatusConflict, err.Error())
		return
	}
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

// releaseKey writes the secrets for the other keys before one goes. The
// last key that can encrypt them cannot go while there are any.
func (h *Handlers) releaseKey(ctx context.Context, name string) error {
	count, err := h.opts.Secrets.Count(ctx)
	if err != nil || count == 0 {
		return err
	}
	vault := h.opts.Secrets.Vault
	if len(vault.Recipients(name)) == 0 && len(vault.Recipients()) > 0 {
		return errors.New("the secrets are encrypted with this key: generate or upload another before deleting it")
	}
	if _, err := h.opts.Secrets.Rekey(ctx, name); err != nil {
		return fmt.Errorf("the secrets cannot be written for the other keys: %w", err)
	}
	return nil
}
