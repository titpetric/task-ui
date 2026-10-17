package api

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ProjectCloneRequest adds a project from a git repository.
type ProjectCloneRequest struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Branch string `json:"branch"`
}

// ProjectRequest changes what the configuration says about a folder.
type ProjectRequest struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Allow       []string            `json:"allow"`
	Commands    map[string]string   `json:"commands"`
	Sources     []string            `json:"sources"`
	JobAllow    map[string][]string `json:"job_allow"`
}

// requireAdmin answers anyone but an admin with 403 and reports it.
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}

// listProjects returns the directories under the root with their
// sources, checkouts and project rows. Admins only.
func (h *Handlers) listProjects(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	entries := folder.Entries(h.opts.Root, h.opts.Git)
	projects, err := h.opts.Projects.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "projects": projects})
}

// folderIDFromURL derives a folder name from a repository URL.
func folderIDFromURL(repo string) string {
	base := strings.TrimSuffix(strings.TrimRight(repo, "/"), ".git")
	if i := strings.LastIndexAny(base, "/:"); i >= 0 {
		base = base[i+1:]
	}
	return folder.SafeID(base)
}

// cloneProject clones a repository into the root and records it. The
// clone is the test: a failure reports git's output and keeps nothing.
func (h *Handlers) cloneProject(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
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
		fail(w, http.StatusBadRequest, "the folder name may hold letters, digits, dots, dashes and underscores")
		return
	}
	dir := filepath.Join(h.opts.Root, req.ID)
	output, err := h.opts.Git.Clone(r.Context(), req.URL, dir, req.Branch)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	project := &model.Project{Folder: req.ID, SourceURL: req.URL, Branch: req.Branch}
	if err := h.opts.Projects.Upsert(r.Context(), project); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"project": project, "runnable": detector.Any(dir), "output": output})
}

// getProject returns the project row of a folder.
func (h *Handlers) getProject(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	project, err := h.opts.Projects.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// updateProject writes the configuration of a folder, creating the row
// when the folder has none.
func (h *Handlers) updateProject(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id := param(r, "folder")
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	var req ProjectRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	project, err := h.opts.Projects.GetByFolder(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		project = &model.Project{Folder: id}
	} else if err != nil {
		failErr(w, err)
		return
	}
	project.Title, project.Description = req.Title, req.Description
	project.Allow = model.EncodeList(req.Allow)
	project.Commands = model.EncodeStrings(req.Commands)
	project.Sources = model.EncodeList(req.Sources)
	project.JobAllow = model.EncodeMap(req.JobAllow)
	if err := h.opts.Projects.Upsert(r.Context(), project); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// deleteProject removes a cloned folder and its row. A folder that is
// not a checkout keeps its files; only the row goes.
func (h *Handlers) deleteProject(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id := param(r, "folder")
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	dir := filepath.Join(h.opts.Root, id)
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
