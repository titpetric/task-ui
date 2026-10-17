package api

import (
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
	writeJSON(w, http.StatusCreated, key)
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
	writeJSON(w, http.StatusCreated, key)
}

// deleteKey removes a pair, or a backup.
func (h *Handlers) deleteKey(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.opts.Keys.Delete(param(r, "name")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
