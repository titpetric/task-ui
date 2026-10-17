package frontend

import (
	"net/http"
	"sort"

	"github.com/titpetric/task-ui/frontend/view"
)

// envFile shows an env file of a folder, to edit. Admins only: the
// content may hold what the folder's own file keeps.
func (h *Handlers) envFile(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	f := h.visibleFolder(r)
	if f == nil {
		h.missing(w, r, "No folder named "+param(r, "folder")+".")
		return
	}
	file, err := h.opts.Env.File(r.Context(), f, param(r, "*"))
	if err != nil {
		h.missing(w, r, err.Error())
		return
	}
	p := h.page(r, f.ID+" / "+file.Path)
	p.Folder, p.EnvFile = f, &file
	p.Crumbs = view.Crumbs(f.ID, "/f/"+f.ID, file.Path)
	if secrets, err := h.opts.Secrets.For(r.Context(), f.ID); err == nil {
		for name := range secrets {
			p.SecretNames = append(p.SecretNames, name)
		}
		sort.Strings(p.SecretNames)
	}
	h.render(w, http.StatusOK, "envfile", p)
}
