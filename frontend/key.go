package frontend

import (
	"net/http"
	"path/filepath"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/sshkey"
)

// keys lists the key pairs of the instance.
func (h *Handlers) keys(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "SSH keys")
	p.Crumbs = view.Crumbs("Settings", "/settings", "SSH keys")
	p.Keys, p.KeysDir = h.opts.Keys.List(), h.opts.Keys.Dir
	p.KeyFolders = h.keyFolders()
	p.KeyPolicies = h.keyPolicies(r)
	p.Encryption = h.opts.Secrets.Vault.Recipients()
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

// keyFolders lists the .ssh/ folders of the projects, the checkouts at
// the top of the root, that hold a key: what their runs load into an ssh
// agent. They are the projects' to manage, so the page only shows them.
func (h *Handlers) keyFolders() []view.KeyFolder {
	var result []view.KeyFolder
	for _, f := range h.opts.Catalogue.List() {
		if f.Depth() != 0 || h.opts.Git == nil || !h.opts.Git.IsRepo(f.Dir) {
			continue
		}
		dir := filepath.Join(f.Dir, ".ssh")
		keys := (&sshkey.Store{Dir: dir}).List()
		if len(keys) == 0 {
			continue
		}
		result = append(result, view.KeyFolder{Project: f.ID, Dir: dir, Keys: keys})
	}
	return result
}

// keyPolicies returns the roles each instance key is given to, none when
// they cannot be read.
func (h *Handlers) keyPolicies(r *http.Request) map[string][]string {
	if h.opts.KeyPolicies == nil {
		return map[string][]string{}
	}
	policies, err := h.opts.KeyPolicies.All(r.Context())
	if err != nil {
		return map[string][]string{}
	}
	return policies
}

// keyEdit shows the form that sets the roles whose runs get a key.
func (h *Handlers) keyEdit(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	name := param(r, "name")
	for _, key := range h.opts.Keys.List() {
		if key.Name != name || key.Backup {
			continue
		}
		p := h.page(r, key.Name)
		p.Crumbs = view.Crumbs("Settings", "/settings", "SSH keys", "/settings/keys", key.Name)
		p.Key, p.KeyPolicies = &key, h.keyPolicies(r)
		h.render(w, http.StatusOK, "key_edit", p)
		return
	}
	h.notFound(w, r)
}
