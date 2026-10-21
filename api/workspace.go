package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// promptJob is the internal job a prompt against a workspace runs.
const promptJob = "claude-prompt"

// listWorkspaces returns every workspace, by folder. Admins only.
func (h *Handlers) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	workspaces, err := h.opts.Workspaces.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	if workspaces == nil {
		workspaces = []model.Workspace{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": workspaces})
}

// createWorkspace clones a repository into <root>/<folder> and records
// it. The folder is the user's to name, one segment that no workspace,
// project or directory under the root holds yet. As with a project, the
// clone is the test: a failure reports git's output and keeps nothing.
func (h *Handlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req WorkspaceRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.URL, req.Folder = strings.TrimSpace(req.URL), strings.TrimSpace(req.Folder)
	if !h.opts.Git.ValidURL(req.URL) {
		fail(w, http.StatusBadRequest, "enter a git URL: https://, ssh://, git@host:path or file://")
		return
	}
	if !folder.ValidID(req.Folder) || strings.Contains(req.Folder, "/") {
		fail(w, http.StatusBadRequest, "name the workspace's folder: letters, digits, dots, dashes and underscores, no slashes")
		return
	}
	if model.IsInternal(req.Folder) {
		fail(w, http.StatusBadRequest, internalRefused)
		return
	}
	ctx := r.Context()
	if _, err := h.opts.Workspaces.GetByFolder(ctx, req.Folder); err == nil {
		fail(w, http.StatusConflict, "a workspace named "+req.Folder+" exists already")
		return
	}
	if _, err := h.opts.Projects.GetByFolder(ctx, req.Folder); err == nil {
		fail(w, http.StatusConflict, "a project named "+req.Folder+" exists already")
		return
	}
	dir := filepath.Join(h.opts.Root, req.Folder)
	if _, err := os.Stat(dir); err == nil {
		fail(w, http.StatusConflict, req.Folder+" exists under the root already")
		return
	}
	output, err := h.opts.Git.Clone(ctx, req.URL, dir, req.Branch)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	workspace := &model.Workspace{
		Folder: req.Folder, Title: req.Title, Description: req.Description,
		SourceURL: req.URL, Branch: req.Branch,
	}
	if err := h.opts.Workspaces.Create(ctx, workspace); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"workspace": workspace, "output": output})
}

// getWorkspace returns the workspace of a folder.
func (h *Handlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

// updateWorkspace writes a workspace's title and description. Its
// folder, repository and branch stay: another checkout is another
// workspace.
func (h *Handlers) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req WorkspaceRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	workspace.Title, workspace.Description = req.Title, req.Description
	if err := h.opts.Workspaces.Update(r.Context(), workspace); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

// deleteWorkspace removes a workspace's checkout, the projects cloned
// inside it with their rows, and its row. A workspace whose folder is
// not a checkout keeps its files; only the rows go.
func (h *Handlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	ctx := r.Context()
	workspace, err := h.opts.Workspaces.GetByFolder(ctx, param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	dir := filepath.Join(h.opts.Root, workspace.Folder)
	if h.opts.Git.IsRepo(dir) {
		if err := h.opts.Git.Remove(dir); err != nil {
			failErr(w, err)
			return
		}
	}
	projects, err := h.opts.Projects.List(ctx)
	if err != nil {
		failErr(w, err)
		return
	}
	for _, project := range projects {
		if !strings.HasPrefix(project.Folder, workspace.Folder+"/") {
			continue
		}
		if err := h.opts.Projects.Delete(ctx, project.ID); err != nil {
			failErr(w, err)
			return
		}
	}
	if err := h.opts.Workspaces.Delete(ctx, workspace.ID); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// promptWorkspace runs a prompt against a workspace: internal/claude-prompt
// as the user, with TASK_UI_PROMPT and the workspace's checkout in
// TASK_UI_WORKSPACE_DIR, which Claude Code works in. It answers with the
// run, for the window to attach to.
func (h *Handlers) promptWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req PromptRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	user := access.UserFrom(r.Context())
	f, job, problem := h.target(r, model.Internal, promptJob)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if !canRun(user, f, job) {
		fail(w, http.StatusForbidden, "your role may view this job, but not run it")
		return
	}
	session, err := h.opts.Runner.Start(r.Context(), runner.Request{
		Folder: f, Job: job, User: user, Cols: req.Cols, Rows: req.Rows,
		Workspace: filepath.Join(h.opts.Root, workspace.Folder),
		Env:       []string{"TASK_UI_PROMPT=" + prompt},
	})
	if errors.Is(err, runner.ErrBusy) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "a prompt is running already", "run": h.view(r, session.Snapshot())})
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, h.view(r, session.Snapshot()))
}

// workspaceFolders returns the folder of every workspace, none without
// the storage or on an error.
func (h *Handlers) workspaceFolders(ctx context.Context) []string {
	if h.opts.Workspaces == nil {
		return nil
	}
	folders, _ := h.opts.Workspaces.Folders(ctx)
	return folders
}
