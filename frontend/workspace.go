package frontend

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/frontend/view"
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
	h.render(w, http.StatusOK, "workspaces", p)
}

// workspaceNew shows the form that clones a workspace.
func (h *Handlers) workspaceNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	p := h.page(r, "Add workspace")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", "Add")
	p.Step = view.StepWorkspaces
	h.render(w, http.StatusOK, "workspace_new", p)
}

// workspaceEdit shows a workspace's title and description, its keys,
// and its delete, which names the projects that go with it.
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
	p := h.page(r, "Configure "+workspace.Folder)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", workspace.Folder)
	p.Step = view.StepWorkspaces
	p.Workspace = workspace
	p.Root = h.opts.Root
	p.Nested = h.nestedProjects(r.Context())
	p.Form = map[string]string{
		"title":       workspace.Title,
		"description": workspace.Description,
	}
	h.workspaceKeys(r, p)
	h.render(w, http.StatusOK, "workspace_edit", p)
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
