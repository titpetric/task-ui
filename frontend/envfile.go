package frontend

import (
	"net/http"
	"sort"

	"github.com/titpetric/task-ui/folder"
)

// envFile shows an env file of a folder, to edit. Admins only: the
// content may hold what the folder's own file keeps.
func (h *Handlers) envFile(w http.ResponseWriter, r *http.Request, f *folder.Folder, path string) {
	if !h.requireAdmin(w, r) {
		return
	}
	file, err := h.opts.Env.File(r.Context(), f, path)
	if err != nil {
		h.missing(w, r, err.Error())
		return
	}
	p := h.page(r, f.ID+" / "+file.Path)
	p.Folder, p.EnvFile = f, &file
	p.Crumbs = crumbsOf(f.ID, file.Path)
	if secrets, err := h.opts.Secrets.For(r.Context(), f.ID); err == nil {
		for name := range secrets {
			p.SecretNames = append(p.SecretNames, name)
		}
		sort.Strings(p.SecretNames)
	}
	h.render(w, http.StatusOK, "envfile", p)
}
