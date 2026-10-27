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
	p.SecretCounts = h.secretCounts(r.Context(), p.WorkspaceList)
	p.Homes = h.homes(p.WorkspaceList)
	h.render(w, http.StatusOK, "workspaces", p)
}

// workspaceNew shows the form that adds a workspace.
func (h *Handlers) workspaceNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	p := h.page(r, "Add workspace")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", "Add workspace")
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
	p := h.page(r, "Edit "+workspace.Folder)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", crumbName(workspace.Folder), view.FolderHref(workspace.Folder, p.Workspaces), "Edit")
	p.Step = view.StepWorkspaces
	p.Workspace = workspace
	p.Root = h.opts.Root
	p.Nested = h.nestedProjects(r.Context())
	p.SecretCounts = h.secretCounts(r.Context(), []model.Workspace{*workspace})
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
// the folders at the top of the root, less the workspaces, for internal.
func (h *Handlers) workspaceProjects(ctx context.Context, p *view.Page, name string) {
	named := h.workspaceFolders(ctx)
	for _, entry := range folder.Entries(h.opts.Root, h.opts.Git, h.workspaceList(ctx)) {
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

// secretCounts returns how many secrets each workspace holds, by its
// folder, none without a secret store.
func (h *Handlers) secretCounts(ctx context.Context, workspaces []model.Workspace) map[string]int {
	counts := map[string]int{}
	if h.opts.Secrets == nil {
		return counts
	}
	for _, ws := range workspaces {
		secrets, _ := h.opts.Secrets.Encrypted(ctx, ws.Folder)
		counts[ws.Folder] = len(secrets)
	}
	return counts
}

// nestedProjects returns the projects inside each workspace, by
// workspace folder: the checkouts under its src and the project rows
// there, the direct ones, each once, in name order. They are what the Projects column
// counts and a delete removes. The default workspace holds the root's.
func (h *Handlers) nestedProjects(ctx context.Context) map[string][]string {
	nested := map[string][]string{}
	named := h.workspaceFolders(ctx)
	add := func(id string) {
		ws, rel := layout.Split(named, id)
		if rel == "" || strings.Contains(rel, "/") || model.IsReserved(rel) || slices.Contains(nested[ws], id) {
			return
		}
		nested[ws] = append(nested[ws], id)
	}
	for _, entry := range folder.Entries(h.opts.Root, h.opts.Git, h.workspaceList(ctx)) {
		add(entry.ID)
	}
	projects, _ := h.opts.Projects.List(ctx)
	for _, project := range projects {
		if !strings.Contains(project.Folder, "/") || slices.Contains(named, strings.SplitN(project.Folder, "/", 2)[0]) {
			add(project.Folder)
		}
	}
	for _, ids := range nested {
		slices.Sort(ids)
	}
	return nested
}
