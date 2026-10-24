package frontend

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// workspaces lists the workspaces and the projects cloned inside each.
func (h *Handlers) workspaces(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	p := h.page(r, "Workspaces")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces")
	p.Step = view.StepWorkspaces
	p.Root = h.opts.Root
	p.WorkspaceList = h.workspaceList(r.Context())
	p.Nested = h.nestedProjects(r.Context())
	p.Homes = h.homes(p.WorkspaceList)
	h.render(w, http.StatusOK, "workspaces", p)
}

// workspaceNew shows the form that adds a workspace.
func (h *Handlers) workspaceNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	p := h.page(r, "Add workspace")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", "Add")
	p.Step = view.StepWorkspaces
	h.render(w, http.StatusOK, "workspace_new", p)
}

// workspaceEdit shows a workspace's settings, its home repository, its
// projects, its keys, and its delete, which names the projects that go
// with it.
func (h *Handlers) workspaceEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	name := param(r, "folder")
	if h.opts.Workspaces == nil {
		h.missing(w, r, "No workspace named "+name+".")
		return
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), name)
	if err != nil || workspace == nil {
		h.missing(w, r, "No workspace named "+name+".")
		return
	}
	p := h.page(r, "Workspace settings: "+workspace.Folder)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", workspace.Folder)
	p.Step = view.StepWorkspaces
	p.Workspace = workspace
	p.Root = h.opts.Root
	p.Nested = h.nestedProjects(r.Context())
	p.Form = map[string]string{
		"title":       workspace.Title,
		"description": workspace.Description,
	}
	h.workspaceProjects(r.Context(), p, workspace.Folder)
	h.workspaceHome(r, p, workspace)
	h.workspaceKeys(r, p)
	h.render(w, http.StatusOK, "workspace_edit", p)
}

// workspaceProjects puts the projects of a workspace on its page, as
// the Projects page lists them: the checkouts inside a named workspace,
// the folders at the top of the root, less the workspaces, for _.
func (h *Handlers) workspaceProjects(ctx context.Context, p *view.Page, name string) {
	named := h.workspaceFolders(ctx)
	for _, entry := range folder.Entries(h.opts.Root, h.opts.Git, named) {
		top, _, nested := strings.Cut(entry.ID, "/")
		if (name == model.Default && !nested && !slices.Contains(named, top)) || (nested && top == name) {
			p.Entries = append(p.Entries, entry)
		}
	}
	h.measure(p)
	h.pulls(p)
}

// workspaceHome puts a workspace's home on its settings page: its
// repository, what git knows about the checkout, and the pull job of
// its git skill when the user may run it.
func (h *Handlers) workspaceHome(r *http.Request, p *view.Page, ws *model.Workspace) {
	home := &view.WorkspaceHome{Folder: layout.HomeID(ws.Folder), Dir: layout.HomeDir(h.opts.Root, ws.Folder), URL: ws.HomeURL, Branch: ws.HomeBranch}
	if h.opts.Git != nil && h.opts.Git.IsRepo(home.Dir) {
		home.Checkout = h.opts.Git.Info(home.Dir)
	}
	if f := h.opts.Catalogue.Find(home.Folder); f != nil {
		if job := pullJob(f.Skills()); job != nil && p.CanRun(f, job) {
			home.Pull = job
		}
	}
	p.Home = home
}

// homes returns the URL of each workspace's home repository, by the
// workspace's folder; an empty home has an empty URL.
func (h *Handlers) homes(workspaces []model.Workspace) map[string]string {
	homes := map[string]string{}
	for _, ws := range workspaces {
		homes[ws.Folder] = ws.HomeURL
	}
	return homes
}

// workspaceList returns every workspace, none without the storage or on
// an error.
func (h *Handlers) workspaceList(ctx context.Context) []model.Workspace {
	if h.opts.Workspaces == nil {
		return nil
	}
	workspaces, _ := h.opts.Workspaces.List(ctx)
	return workspaces
}

// nestedProjects returns the folders of the project rows inside each
// workspace, by workspace folder: the projects a delete removes. The
// default workspace holds the rows at the top of the root.
func (h *Handlers) nestedProjects(ctx context.Context) map[string][]string {
	nested := map[string][]string{}
	projects, _ := h.opts.Projects.List(ctx)
	named := h.workspaceFolders(ctx)
	for _, project := range projects {
		top, _, _ := strings.Cut(project.Folder, "/")
		if !slices.Contains(named, top) && !strings.Contains(project.Folder, "/") && !model.IsReserved(top) {
			nested[model.Default] = append(nested[model.Default], project.Folder)
		}
	}
	for _, ws := range named {
		for _, project := range projects {
			if strings.HasPrefix(project.Folder, ws+"/") {
				nested[ws] = append(nested[ws], project.Folder)
			}
		}
	}
	return nested
}
