package frontend

import (
	"net/http"
	"sort"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
)

// envFile shows an env file a user filled for a folder, to edit, or with
// no path the form that adds one. Takes env/view, and env/edit to add.
func (h *Handlers) envFile(w http.ResponseWriter, r *http.Request, f *folder.Folder, path string) {
	if !h.require(w, r, "env/view") {
		return
	}
	file := envfile.File{Folder: f.ID}
	title := "Add env file"
	if path != "" {
		var err error
		if file, err = h.opts.Env.File(r.Context(), f, path); err != nil {
			h.missing(w, r, err.Error())
			return
		}
		title = file.Path
	} else if !h.require(w, r, "env/edit") {
		return
	}
	p := h.page(r, f.ID+" / "+title)
	p.Folder, p.EnvFile = f, &file
	p.Crumbs = crumbsOf(f.ID, title, p.Workspaces)
	if secrets, err := h.opts.Secrets.For(r.Context(), f.ID); err == nil {
		for name := range secrets {
			p.SecretNames = append(p.SecretNames, name)
		}
		sort.Strings(p.SecretNames)
	}
	h.render(w, http.StatusOK, "envfile", p)
}
