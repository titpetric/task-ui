package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workspace"
)

// listWorkspaces returns every workspace, the default first, then in
// their order. Takes workspace/view.
func (h *Handlers) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
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

// createWorkspace makes the directory <root>/<folder> and records it.
// The folder is the user's to name, one segment that no workspace,
// project or directory under the root holds yet. With a URL it clones
// the repository into the workspace as a project, the folder the
// request names or the repository's, and makes it the home project. As
// with a project, the clone is the test: a failure reports git's output
// and keeps nothing.
func (h *Handlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	var req WorkspaceRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.URL, req.Folder, req.Project = strings.TrimSpace(req.URL), strings.TrimSpace(req.Folder), strings.TrimSpace(req.Project)
	if req.URL != "" && !h.opts.Git.ValidURL(req.URL) {
		fail(w, http.StatusBadRequest, "enter a git URL: https://, ssh://, git@host:path or file://")
		return
	}
	if req.URL != "" && req.Project == "" {
		req.Project = folderIDFromURL(req.URL)
	}
	if req.URL != "" && (!folder.ValidID(req.Project) || strings.Contains(req.Project, "/") || model.IsReserved(req.Project)) {
		fail(w, http.StatusBadRequest, "name the home project's folder: letters, digits, dots, dashes and underscores, no slashes")
		return
	}
	ctx := r.Context()
	if status, problem := h.freeWorkspace(ctx, req.Folder); problem != "" {
		fail(w, status, problem)
		return
	}
	dir := filepath.Join(h.opts.Root, req.Folder)
	if err := os.Mkdir(dir, 0o755); err != nil {
		failErr(w, err)
		return
	}
	output := ""
	if req.URL != "" {
		out, err := h.opts.Git.Clone(ctx, req.URL, filepath.Join(dir, req.Project), req.Branch)
		if err != nil {
			os.RemoveAll(dir)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": out})
			return
		}
		output = out
	}
	workspace := &model.Workspace{
		Folder: req.Folder, Title: req.Title, Description: req.Description,
		Shallow: req.Recursive != nil && !*req.Recursive,
	}
	project, err := h.recordWorkspace(ctx, workspace, req)
	if err != nil {
		os.RemoveAll(dir)
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"workspace": workspace, "project": project, "output": output})
}

// recordWorkspace adds a new workspace's row, and with a URL the row of
// its home project, nil without. A failure keeps no row.
func (h *Handlers) recordWorkspace(ctx context.Context, workspace *model.Workspace, req WorkspaceRequest) (*model.Project, error) {
	if err := h.opts.Workspaces.Create(ctx, workspace); err != nil {
		return nil, err
	}
	if req.URL == "" {
		return nil, nil
	}
	project := &model.Project{Folder: workspace.Folder + "/" + req.Project, SourceURL: req.URL, Branch: req.Branch}
	if err := h.opts.Projects.Upsert(ctx, project); err != nil {
		_ = h.opts.Workspaces.Delete(context.WithoutCancel(ctx), workspace.ID)
		return nil, err
	}
	workspace.HomeProjectID = project.ID
	if err := h.opts.Workspaces.Update(ctx, workspace); err != nil {
		_ = h.opts.Projects.Delete(context.WithoutCancel(ctx), project.ID)
		_ = h.opts.Workspaces.Delete(context.WithoutCancel(ctx), workspace.ID)
		return nil, err
	}
	return project, nil
}

// errHome refuses a home project that is no project at the top of the
// workspace, a 400.
var errHome = errors.New("the home project is a project at the top of its workspace")

// homeProject returns the ID of the project a request names as a
// workspace's home, by its folder: "" for "", errHome for a folder that
// is no project at the top of the workspace.
func (h *Handlers) homeProject(ctx context.Context, ws *model.Workspace, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}
	relative := workspace.Relative(ws.Folder, name)
	if relative == "" || strings.Contains(relative, "/") || !workspace.Contains(ws.Folder, h.workspaceFolders(ctx), name) {
		return "", errHome
	}
	project, err := h.opts.Projects.GetByFolder(ctx, name)
	if errors.Is(err, storage.ErrNotFound) {
		return "", errHome
	}
	if err != nil {
		return "", err
	}
	return project.ID, nil
}

// setHome makes a project the home project of the workspace of a
// folder, _ for the default workspace.
func (h *Handlers) setHome(ctx context.Context, folder, projectID string) error {
	if folder == "" {
		folder = model.Default
	}
	ws, err := h.opts.Workspaces.GetByFolder(ctx, folder)
	if err != nil {
		return err
	}
	ws.HomeProjectID = projectID
	return h.opts.Workspaces.Update(ctx, ws)
}

// freeWorkspace says why a new workspace may not take a folder, with
// the status to answer: 400 for a name that is not one segment or that
// task-ui keeps, 409 for a folder a workspace, a project or a directory
// under the root holds. "" for a free folder.
func (h *Handlers) freeWorkspace(ctx context.Context, name string) (int, string) {
	if !folder.ValidID(name) || strings.Contains(name, "/") {
		return http.StatusBadRequest, "name the workspace's folder: letters, digits, dots, dashes and underscores, no slashes"
	}
	if model.IsReserved(name) {
		return http.StatusBadRequest, internalRefused
	}
	if _, err := h.opts.Workspaces.GetByFolder(ctx, name); err == nil {
		return http.StatusConflict, "a workspace named " + name + " exists already"
	}
	if _, err := h.opts.Projects.GetByFolder(ctx, name); err == nil {
		return http.StatusConflict, "a project named " + name + " exists already"
	}
	if _, err := os.Stat(filepath.Join(h.opts.Root, name)); err == nil {
		return http.StatusConflict, name + " exists under the root already"
	}
	return 0, ""
}

// getWorkspace returns the workspace of a folder.
func (h *Handlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

// updateWorkspace writes a workspace's title, description, Recursive
// and home project. Its folder stays. The default workspace keeps its
// place, first. A home project that is no project at the top of the
// workspace is a 400.
func (h *Handlers) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
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
	if req.Sort != nil && workspace.Folder != model.Default {
		workspace.Sort = *req.Sort
	}
	if req.Recursive != nil {
		workspace.Shallow = !*req.Recursive
	}
	if req.HomeProject != nil {
		id, err := h.homeProject(r.Context(), workspace, *req.HomeProject)
		if errors.Is(err, errHome) {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err != nil {
			failErr(w, err)
			return
		}
		workspace.HomeProjectID = id
	}
	if err := h.opts.Workspaces.Update(r.Context(), workspace); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

// deleteWorkspace removes the projects cloned inside a workspace, their
// checkouts and rows, its secrets, its row, its keys with their grants,
// and its directory once empty: what else it holds stays. The default
// workspace, the root, stays.
func (h *Handlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/delete") {
		return
	}
	ctx := r.Context()
	workspace, err := h.opts.Workspaces.GetByFolder(ctx, param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	if workspace.Folder == model.Default {
		fail(w, http.StatusConflict, "you can't remove the default workspace: it holds the projects at the root")
		return
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
		if dir := filepath.Join(h.opts.Root, filepath.FromSlash(project.Folder)); h.opts.Git.IsRepo(dir) {
			if err := h.opts.Git.Remove(dir); err != nil {
				failErr(w, err)
				return
			}
		}
		if err := h.opts.Projects.Delete(ctx, project.ID); err != nil {
			failErr(w, err)
			return
		}
	}
	if h.opts.Secrets != nil {
		if _, err := h.opts.Secrets.Forget(ctx, workspace.Folder); err != nil {
			failErr(w, err)
			return
		}
	}
	if err := h.forgetEnvFiles(ctx, workspace.Folder); err != nil {
		failErr(w, err)
		return
	}
	if err := h.opts.Workspaces.Delete(ctx, workspace.ID); err != nil {
		failErr(w, err)
		return
	}
	if h.opts.Keys != nil {
		if err := h.opts.Keys.RemoveScope(ctx, keystore.ScopeWorkspace(workspace.ID)); err != nil {
			failErr(w, err)
			return
		}
	}
	_ = os.Remove(workspace.Dir(h.opts.Root))
	w.WriteHeader(http.StatusNoContent)
}

// promptWorkspace puts a prompt to Claude in a workspace, as POST
// /api/prompts does in a folder, and answers the same: the note, which
// the board answers by running @claude-prompt there as the user. Claude
// Code works in the home project, TASK_UI_HOME, or else the workspace's
// directory, TASK_UI_WORKSPACE_DIR.
func (h *Handlers) promptWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/edit") || !h.board(w) {
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
	h.ask(w, r, req, workspace.Folder)
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

// forgetEnvFiles deletes the managed env files of a named workspace's
// folders: the workspace's own and its projects'.
func (h *Handlers) forgetEnvFiles(ctx context.Context, name string) error {
	if h.opts.Env == nil {
		return nil
	}
	files, err := h.opts.Env.Managed.All(ctx)
	if err != nil {
		return err
	}
	for _, file := range files {
		if top, _, _ := strings.Cut(file.Folder, "/"); top != name {
			continue
		}
		if err := h.opts.Env.Managed.Delete(ctx, file.ID); err != nil {
			return err
		}
	}
	return nil
}
