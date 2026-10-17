package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// KeyRequest generates a key pair, or uploads a private key.
type KeyRequest struct {
	// Type is the pair to generate: ed25519, ecdsa or rsa.
	Type    string `json:"type"`
	Comment string `json:"comment"`
	// Content is the private key to upload, PEM encoded.
	Content string `json:"content"`
}

// listKeys returns the key pairs of the instance SSH folder. Admins only.
func (h *Handlers) listKeys(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": h.opts.Keys.List(), "dir": h.opts.Keys.Dir})
}

// generateKey makes a new pair. The pair of the same type that was there
// is kept as a backup.
func (h *Handlers) generateKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req KeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	key, err := h.opts.Keys.Generate(req.Type, strings.TrimSpace(req.Comment))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	h.rekeyed(w, r, http.StatusCreated, key)
}

// rekeyed answers a key change with the key and how many secrets were
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
	key, err := h.opts.Keys.Upload([]byte(req.Content), strings.TrimSpace(req.Comment))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	h.rekeyed(w, r, http.StatusCreated, key)
}

// deleteKey removes a pair, or a backup. The secrets are written again
// for the keys that remain first; the last key that can encrypt them
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
