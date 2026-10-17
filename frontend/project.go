package frontend

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
)

// projects lists the directories under the root.
func (h *Handlers) projects(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Projects")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects")
	p.Root = h.opts.Root
	p.Entries = folder.Entries(h.opts.Root, h.opts.Git)
	h.render(w, http.StatusOK, "projects", p)
}

// projectNew shows the clone form.
func (h *Handlers) projectNew(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Add project")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects", "/settings/projects", "Add")
	h.render(w, http.StatusOK, "project_new", p)
}

// projectEdit shows the configuration of a folder.
func (h *Handlers) projectEdit(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id := param(r, "folder")
	f := h.opts.Catalogue.Find(id)
	if f == nil {
		if !folder.ValidID(id) {
			h.missing(w, r, "No folder named "+id+".")
			return
		}
		f = &folder.Folder{ID: id, Title: id}
	}
	p := h.page(r, "Configure "+id)
	p.Folder = f
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects", "/settings/projects", id)
	if project, err := h.opts.Projects.GetByFolder(r.Context(), id); err == nil {
		p.Form = map[string]string{
			"title":       project.Title,
			"description": project.Description,
			"allow":       strings.Join(project.AllowList(), ", "),
			"sources":     strings.Join(project.SourceList(), ", "),
			"hours":       project.Hours,
		}
	}
	h.render(w, http.StatusOK, "project_edit", p)
}
