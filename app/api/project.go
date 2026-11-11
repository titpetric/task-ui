package api

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ProjectCloneRequest adds a project from a git repository.
type ProjectCloneRequest struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Branch string `json:"branch"`
	// Workspace clones the project inside that workspace's directory:
	// the folder becomes <workspace>/<id>. Empty or internal, the default
	// workspace, clones into the root.
	Workspace string `json:"workspace,omitempty"`
	// Recursive lists the folders inside the checkout as deep as they go,
	// the default; false lists the project and its top-level folders alone.
	Recursive *bool `json:"recursive,omitempty"`
	// Sidebar lists the project in the sidebar, the default; false keeps
	// it, and the folders inside it, out. Its pages stay.
	Sidebar *bool `json:"sidebar,omitempty"`
}

// ProjectRequest changes what the configuration says about a folder.
type ProjectRequest struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Allow       []string            `json:"allow"`
	Commands    map[string]string   `json:"commands"`
	Sources     []string            `json:"sources"`
	JobAllow    map[string][]string `json:"job_allow"`
	// Recursive lists the folders inside the project as deep as they go,
	// the default; false lists the project and its top-level folders alone.
	// Left out, it stays unchanged.
	Recursive *bool `json:"recursive,omitempty"`
	// Sidebar lists the project in the sidebar, the default; false keeps
	// it, and the folders inside it, out. Left out, it stays unchanged.
	Sidebar *bool `json:"sidebar,omitempty"`
	// Workspace moves the project into that workspace, internal for the root:
	// the checkout and the row take the ID <workspace>/<last segment>,
	// or the last segment alone in internal. Left out, or naming the workspace
	// that holds it, the project stays.
	Workspace *string `json:"workspace,omitempty"`
}

// apply writes the request onto the row.
func (req ProjectRequest) apply(project *model.Project) {
	project.Title, project.Description = req.Title, req.Description
	project.Allow = model.EncodeList(req.Allow)
	project.Commands = model.EncodeStrings(req.Commands)
	project.Sources = model.EncodeList(req.Sources)
	project.JobAllow = model.EncodeMap(req.JobAllow)
	if req.Recursive != nil {
		project.Shallow = !*req.Recursive
	}
	if req.Sidebar != nil {
		project.SidebarHidden = !*req.Sidebar
	}
}

// listProjects returns the directories under the root with their
// sources, checkouts and project rows. Takes project/view.
func (h *Handlers) listProjects(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/view") {
		return
	}
	entries := folder.Entries(h.opts.Root, h.opts.Git, h.workspaceRows(r.Context()))
	projects, err := h.opts.Projects.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "projects": projects})
}

// reservedRefused refuses a project or a workspace named internal, _,
// home or src, task-ui's own.
const reservedRefused = "internal, _, home and src are task-ui's own names: name the folder otherwise"

// folderIDFromURL derives a folder name from a repository URL.
func folderIDFromURL(repo string) string {
	base := strings.TrimSuffix(strings.TrimRight(repo, "/"), ".git")
	if i := strings.LastIndexAny(base, "/:"); i >= 0 {
		base = base[i+1:]
	}
	return folder.SafeID(base)
}

// cloneProject clones a repository into the src of the default
// workspace, or of a named one, and records it. The clone is the test: a
// failure reports git's output and keeps nothing.
func (h *Handlers) cloneProject(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/edit") {
		return
	}
	var req ProjectCloneRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.URL, req.ID = strings.TrimSpace(req.URL), strings.TrimSpace(req.ID)
	if req.ID == "" {
		req.ID = folderIDFromURL(req.URL)
	}
	if !h.opts.Git.ValidURL(req.URL) {
		fail(w, http.StatusBadRequest, "enter a git URL: https://, ssh://, git@host:path or file://")
		return
	}
	if !folder.ValidID(req.ID) {
		fail(w, http.StatusBadRequest, "a folder name can use only letters, digits, dots, dashes and underscores")
		return
	}
	req.Workspace = model.Canonical(strings.TrimSpace(req.Workspace))
	if req.Workspace == "" {
		req.Workspace = model.Default
	}
	if top, _, _ := strings.Cut(req.ID, "/"); layout.Reserved(top) {
		fail(w, http.StatusBadRequest, reservedRefused)
		return
	}
	id := req.ID
	if _, err := h.opts.Workspaces.GetByFolder(r.Context(), req.Workspace); err != nil && req.Workspace == model.Default {
		fail(w, http.StatusBadRequest, "pick a workspace: the instance has no default workspace")
		return
	} else if err != nil {
		fail(w, http.StatusBadRequest, "there is no workspace named "+req.Workspace)
		return
	}
	if req.Workspace != model.Default {
		id = req.Workspace + "/" + req.ID
	} else if problem := h.rootTaken(r, id); problem != "" {
		fail(w, http.StatusConflict, problem)
		return
	}
	dir := h.dirOf(r.Context(), id)
	output, err := h.opts.Git.Clone(r.Context(), req.URL, dir, req.Branch)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	project := &model.Project{
		Folder: id, SourceURL: req.URL, Branch: req.Branch,
		Shallow:       req.Recursive != nil && !*req.Recursive,
		SidebarHidden: req.Sidebar != nil && !*req.Sidebar,
	}
	if err := h.opts.Projects.Upsert(r.Context(), project); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"project": project, "runnable": detector.Any(dir), "output": output})
}

// rootTaken says why a project cloned into the default workspace may
// not take a folder: a workspace holds its first segment, or a directory
// is there. "" for a free folder.
func (h *Handlers) rootTaken(r *http.Request, id string) string {
	top, _, _ := strings.Cut(id, "/")
	if _, err := h.opts.Workspaces.GetByFolder(r.Context(), top); err == nil {
		return "a workspace named " + top + " exists already: clone into it, or name the folder otherwise"
	}
	if _, err := os.Stat(layout.ProjectDir(h.opts.Root, model.Default, id)); err == nil {
		return id + " exists under the root already"
	}
	return ""
}

// getProject returns the project row of a folder.
func (h *Handlers) getProject(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/view") {
		return
	}
	project, err := h.opts.Projects.GetByFolder(r.Context(), folderParam(r))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// updateProject writes the configuration of a folder, creating the row
// when the folder has none.
func (h *Handlers) updateProject(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/edit") {
		return
	}
	id := folderParam(r)
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	if model.IsReserved(id) || layout.IsHome(h.workspaceFolders(r.Context()), id) {
		fail(w, http.StatusBadRequest, reservedRefused)
		return
	}
	var req ProjectRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	to := id
	if req.Workspace != nil {
		var status int
		var problem string
		if to, status, problem = h.moveTarget(r, id, strings.TrimSpace(*req.Workspace)); problem != "" {
			fail(w, status, problem)
			return
		}
	}
	project, err := h.opts.Projects.GetByFolder(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		project = &model.Project{Folder: id}
	} else if err != nil {
		failErr(w, err)
		return
	}
	req.apply(project)
	if to != id {
		h.moveProject(w, r, project, to)
		return
	}
	if err := h.opts.Projects.Upsert(r.Context(), project); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// deleteProject removes a cloned folder and its row. A folder that is
// not a checkout keeps its files; only the row goes.
func (h *Handlers) deleteProject(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/delete") {
		return
	}
	id := folderParam(r)
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	if layout.IsHome(h.workspaceFolders(r.Context()), id) {
		fail(w, http.StatusBadRequest, "a workspace's home is no project: its settings page replaces or removes it")
		return
	}
	dir := h.dirOf(r.Context(), id)
	if h.opts.Git.IsRepo(dir) {
		if err := h.opts.Git.Remove(dir); err != nil {
			failErr(w, err)
			return
		}
	}
	if project, err := h.opts.Projects.GetByFolder(r.Context(), id); err == nil {
		if err := h.opts.Projects.Delete(r.Context(), project.ID); err != nil {
			failErr(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
