package api

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/hours"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ProjectCloneRequest adds a project from a git repository.
type ProjectCloneRequest struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Branch string `json:"branch"`
	// Recursive lists the folders inside the checkout too, the default;
	// false lists the project's own jobs alone.
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
	// Hours says when the folder's jobs may run, as the hours package
	// reads them; JobHours does so per job.
	Hours    string            `json:"hours"`
	JobHours map[string]string `json:"job_hours"`
	// Recursive lists the folders inside the project too, the default;
	// false lists its own jobs alone. Left out, it stays unchanged.
	Recursive *bool `json:"recursive,omitempty"`
	// Sidebar lists the project in the sidebar, the default; false keeps
	// it, and the folders inside it, out. Left out, it stays unchanged.
	Sidebar *bool `json:"sidebar,omitempty"`
}

// validHours checks the hours of the folder and of its jobs.
func (req ProjectRequest) validHours() error {
	if _, err := hours.Parse(req.Hours); err != nil {
		return err
	}
	for job, text := range req.JobHours {
		if _, err := hours.Parse(text); err != nil {
			return fmt.Errorf("%s: %w", job, err)
		}
	}
	return nil
}

// apply writes the request onto the row.
func (req ProjectRequest) apply(project *model.Project) {
	project.Title, project.Description = req.Title, req.Description
	project.Allow = model.EncodeList(req.Allow)
	project.Commands = model.EncodeStrings(req.Commands)
	project.Sources = model.EncodeList(req.Sources)
	project.JobAllow = model.EncodeMap(req.JobAllow)
	project.Hours = strings.TrimSpace(req.Hours)
	project.JobHours = model.EncodeStrings(req.JobHours)
	if req.Recursive != nil {
		project.RootOnly = !*req.Recursive
	}
	if req.Sidebar != nil {
		project.SidebarHidden = !*req.Sidebar
	}
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

// internalRefused refuses a project in the internal namespace.
const internalRefused = "internal is task-ui's own namespace: name the folder otherwise"

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
	if model.IsInternal(req.ID) {
		fail(w, http.StatusBadRequest, internalRefused)
		return
	}
	dir := filepath.Join(h.opts.Root, req.ID)
	output, err := h.opts.Git.Clone(r.Context(), req.URL, dir, req.Branch)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	project := &model.Project{
		Folder: req.ID, SourceURL: req.URL, Branch: req.Branch,
		RootOnly:      req.Recursive != nil && !*req.Recursive,
		SidebarHidden: req.Sidebar != nil && !*req.Sidebar,
	}
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
	if !requireAdmin(w, r) {
		return
	}
	id := folderParam(r)
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	if model.IsInternal(id) {
		fail(w, http.StatusBadRequest, internalRefused)
		return
	}
	var req ProjectRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.validHours(); err != nil {
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
	req.apply(project)
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
	id := folderParam(r)
	if !folder.ValidID(id) {
		fail(w, http.StatusBadRequest, "invalid folder name")
		return
	}
	dir := filepath.Join(h.opts.Root, filepath.FromSlash(id))
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
