package frontend

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// projects lists the directories under the root, a block per
// workspace.
func (h *Handlers) projects(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/view") {
		return
	}
	p := h.page(r, "Projects")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects")
	p.Step = view.StepProjects
	p.Root = h.opts.Root
	workspaces := h.workspaceList(r.Context())
	p.Entries = folder.Entries(h.opts.Root, h.opts.Git, workspaces)
	p.ProjectGroups = view.ProjectGroups(p.Entries, workspaces)
	h.measure(p)
	h.pulls(p)
	h.render(w, http.StatusOK, "projects", p)
}

// pulls puts the pull job of each entry of a page on it, for the
// entries the catalogue holds whose git skill has one and that the user
// may run: Pull now in the row's menu.
func (h *Handlers) pulls(p *view.Page) {
	for _, entry := range p.Entries {
		f := h.opts.Catalogue.Find(entry.ID)
		if f == nil {
			continue
		}
		job := pullJob(f.Skills())
		if job == nil || !p.CanRun(f, job) {
			continue
		}
		if p.Pulls == nil {
			p.Pulls = map[string]*folder.Job{}
		}
		p.Pulls[entry.ID] = job
	}
}

// measure puts the size on disk of each entry of a page on it, nothing
// without a measure.
func (h *Handlers) measure(p *view.Page) {
	if h.opts.Disk == nil {
		return
	}
	p.Spaces = map[string]diskusage.Usage{}
	for _, entry := range p.Entries {
		p.Spaces[entry.ID] = h.opts.Disk.Measure(entry.Dir)
	}
}

// projectNew shows the clone form.
func (h *Handlers) projectNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/edit") {
		return
	}
	p := h.page(r, "Add project")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects", "/settings/projects", "Add project")
	p.Step = view.StepProjects
	p.WorkspaceList = h.workspaceList(r.Context())
	p.Form = map[string]string{"workspace": r.URL.Query().Get("workspace")}
	h.render(w, http.StatusOK, "project_new", p)
}

// projectEdit shows the configuration of a folder.
func (h *Handlers) projectEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/view") {
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
	p := h.page(r, "Edit "+id)
	p.Folder = f
	p.Crumbs = view.Crumbs("Settings", "/settings", "Projects", "/settings/projects", id, view.FolderHref(id, p.Workspaces), "Edit")
	p.Step = view.StepProjects
	if h.opts.Git != nil {
		p.Checkout = h.opts.Git.Info(layout.Dir(h.opts.Root, h.workspaceFolders(r.Context()), id))
	}
	if project, err := h.opts.Projects.GetByFolder(r.Context(), id); err == nil {
		p.Project = project
		p.Form = map[string]string{
			"title":       project.Title,
			"description": project.Description,
			"sources":     strings.Join(project.SourceList(), ", "),
		}
	}
	p.Users, _ = h.opts.Users.List(r.Context())
	p.Allow, p.JobAllow = p.Project.AllowList(), p.Project.JobAllowMap()
	h.moveChoices(r.Context(), p, id)
	h.render(w, http.StatusOK, "project_edit", p)
}

// moveChoices fills in what the Workspace select of a configure page
// needs: the workspaces, the one the project is in, and the rows a move
// leaves behind. A workspace, or a folder inside a project, gets none.
func (h *Handlers) moveChoices(ctx context.Context, p *view.Page, id string) {
	workspaces := h.workspaceFolders(ctx)
	if slices.Contains(workspaces, id) {
		return
	}
	from := model.Default
	if top, rest, nested := strings.Cut(id, "/"); nested {
		if !slices.Contains(workspaces, top) || strings.Contains(rest, "/") {
			return
		}
		from = top
	}
	p.MoveFrom = from
	p.WorkspaceList = h.workspaceList(ctx)
	if h.opts.FolderRefs == nil {
		return
	}
	refs, _ := h.opts.FolderRefs.Under(ctx, id)
	for _, ref := range refs {
		p.MoveLeft = append(p.MoveLeft, ref.String())
	}
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
