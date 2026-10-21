package frontend

import (
	"context"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/diskusage"
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
	p.Entries = folder.Entries(h.opts.Root, h.opts.Git, h.workspaceFolders(r.Context()))
	if h.opts.Disk != nil {
		p.Spaces = map[string]diskusage.Usage{}
		for _, entry := range p.Entries {
			p.Spaces[entry.ID] = h.opts.Disk.Measure(entry.Dir)
		}
	}
	h.render(w, http.StatusOK, "projects", p)
}

// projectNew shows the clone form.
func (h *Handlers) projectNew(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Add project")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects", "/settings/projects", "Add")
	p.WorkspaceList = h.workspaceList(r.Context())
	h.render(w, http.StatusOK, "project_new", p)
}

// projectEdit shows the configuration of a folder.
func (h *Handlers) projectEdit(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id := strings.Trim(param(r, "*"), "/")
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
		p.Project = project
		p.Form = map[string]string{
			"title":       project.Title,
			"description": project.Description,
			"sources":     strings.Join(project.SourceList(), ", "),
			"hours":       project.Hours,
		}
	}
	users, _ := h.opts.Users.List(r.Context())
	p.AllowChoices = view.WhoChoices(p.Project.AllowList(), users)
	p.JobAllowChoices = map[string][]view.WhoChoice{}
	jobAllow := p.Project.JobAllowMap()
	for _, job := range f.Jobs {
		p.JobAllowChoices[job.ID] = view.WhoChoices(jobAllow[job.ID], users)
	}
	h.render(w, http.StatusOK, "project_edit", p)
}

// workspaceFolders returns the folder of every named workspace, none
// without the storage or on an error.
func (h *Handlers) workspaceFolders(ctx context.Context) []string {
	if h.opts.Workspaces == nil {
		return nil
	}
	folders, _ := h.opts.Workspaces.Folders(ctx)
	return folders
}
