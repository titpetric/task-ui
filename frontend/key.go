package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
)

// keys lists the key pairs of the instance.
func (h *Handlers) keys(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "SSH keys")
	p.Crumbs = view.Crumbs("Settings", "/settings", "SSH keys")
	p.Keys, p.KeysDir = h.opts.Keys.List(), h.opts.Keys.Dir
	h.render(w, http.StatusOK, "keys", p)
}

// keyGenerate shows the generate form.
func (h *Handlers) keyGenerate(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Generate key")
	p.Crumbs = view.Crumbs("Settings", "/settings", "SSH keys", "/settings/keys", "Generate")
	p.KeysDir = h.opts.Keys.Dir
	h.render(w, http.StatusOK, "key_generate", p)
}

// keyUpload shows the upload form.
func (h *Handlers) keyUpload(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Upload key")
	p.Crumbs = view.Crumbs("Settings", "/settings", "SSH keys", "/settings/keys", "Upload")
	p.KeysDir = h.opts.Keys.Dir
	h.render(w, http.StatusOK, "key_upload", p)
}
