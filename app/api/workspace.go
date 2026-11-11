package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
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

// createWorkspace makes the directory <root>/<folder> with its home and
// src and records it. The folder is the user's to name, one segment
// that no workspace, project or directory under the root holds yet.
// With a URL it clones the home repository into the home. As with a
// project, the clone is the test: a failure reports git's output and
// keeps nothing.
func (h *Handlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	var req WorkspaceRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.URL, req.Folder, req.Branch = strings.TrimSpace(req.URL), strings.TrimSpace(req.Folder), strings.TrimSpace(req.Branch)
	if req.URL != "" && !h.opts.Git.ValidURL(req.URL) {
		fail(w, http.StatusBadRequest, "enter a git URL: https://, ssh://, git@host:path or file://")
		return
	}
	hours, _, err := req.hours()
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()
	if status, problem := h.freeWorkspace(ctx, req.Folder); problem != "" {
		fail(w, status, problem)
		return
	}
	dir := layout.WorkspaceDir(h.opts.Root, req.Folder)
	if err := os.Mkdir(dir, 0o755); err != nil {
		failErr(w, err)
		return
	}
	output, err := h.makeHome(ctx, req.Folder, req.URL, req.Branch)
	if err != nil {
		os.RemoveAll(dir)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	workspace := &model.Workspace{
		Folder: req.Folder, Title: req.Title, Description: req.Description,
		Shallow: req.Recursive != nil && !*req.Recursive,
		HomeURL: req.URL, HomeBranch: req.Branch, Hours: hours,
	}
	if err := h.opts.Workspaces.Create(ctx, workspace); err != nil {
		os.RemoveAll(dir)
		failErr(w, err)
		return
	}
	// folder is at the top for the form's data-next, which fills {folder}.
	writeJSON(w, http.StatusCreated, map[string]any{"folder": workspace.Folder, "workspace": workspace, "output": output})
}

// makeHome makes the home and src of a workspace and, with a URL,
// clones the home repository into the home, on branch when named. It
// returns git's output; a clone that fails leaves an empty home.
func (h *Handlers) makeHome(ctx context.Context, name, url, branch string) (string, error) {
	if err := layout.Ensure(h.opts.Root, name); err != nil {
		return "", err
	}
	if url == "" {
		return "", nil
	}
	home := layout.HomeDir(h.opts.Root, name)
	if err := os.Remove(home); err != nil {
		return "", err
	}
	output, err := h.opts.Git.Clone(ctx, url, home, branch)
	if err != nil {
		_ = os.MkdirAll(home, 0o755)
		return output, err
	}
	return output, nil
}

// freeWorkspace says why a new workspace may not take a folder, with
// the status to answer: 400 for a name that is not one segment or that
// task-ui keeps, 409 for a folder a workspace, a project or a directory
// under the root holds. "" for a free folder.
func (h *Handlers) freeWorkspace(ctx context.Context, name string) (int, string) {
	if !folder.ValidID(name) || strings.Contains(name, "/") {
		return http.StatusBadRequest, "name the workspace's folder: letters, digits, dots, dashes and underscores, no slashes"
	}
	if layout.Reserved(name) {
		return http.StatusBadRequest, reservedRefused
	}
	if _, err := h.opts.Workspaces.GetByFolder(ctx, name); err == nil {
		return http.StatusConflict, "a workspace named " + name + " exists already"
	}
	if _, err := h.opts.Projects.GetByFolder(ctx, name); err == nil {
		return http.StatusConflict, "a project named " + name + " exists already"
	}
	if _, err := os.Stat(layout.WorkspaceDir(h.opts.Root, name)); err == nil {
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
// and sidebar limit, a 400 below 0. Its folder and its home stay. The
// default workspace keeps its place, first.
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
	if req.SidebarLimit != nil && *req.SidebarLimit < 0 {
		fail(w, http.StatusBadRequest, "sidebar_limit takes 0, every project, or how many to list")
		return
	}
	text, setHours, err := req.hours()
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	workspace.Title, workspace.Description = req.Title, req.Description
	if setHours {
		workspace.Hours = text
	}
	if req.Sort != nil && workspace.Folder != model.Default {
		workspace.Sort = *req.Sort
	}
	if req.Recursive != nil {
		workspace.Shallow = !*req.Recursive
	}
	if req.SidebarLimit != nil {
		workspace.SidebarLimit = *req.SidebarLimit
	}
	if err := h.opts.Workspaces.Update(r.Context(), workspace); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

// replaceHome replaces the home of a workspace with a clone of the
// repository the request names, the default workspace's too, and
// records the URL and the branch. It refuses, a 409, while a run of the
// workspace is active: the home is the HOME of every run in it. A clone
// that fails leaves an empty home and reports git's output. Takes
// workspace/edit.
func (h *Handlers) replaceHome(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	var req WorkspaceHomeRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.URL, req.Branch = strings.TrimSpace(req.URL), strings.TrimSpace(req.Branch)
	if !h.opts.Git.ValidURL(req.URL) {
		fail(w, http.StatusBadRequest, "enter a git URL: https://, ssh://, git@host:path or file://")
		return
	}
	h.setHome(w, r, req.URL, req.Branch)
}

// removeHome empties the home of a workspace and forgets its repository,
// under the same refusal as replaceHome. Takes workspace/edit.
func (h *Handlers) removeHome(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	h.setHome(w, r, "", "")
}

// setHome removes a workspace's home, clones url into its place when
// named, and writes the row: the one step of Replace and of Remove.
func (h *Handlers) setHome(w http.ResponseWriter, r *http.Request, url, branch string) {
	ctx := r.Context()
	workspace, err := h.opts.Workspaces.GetByFolder(ctx, param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	if busy, err := h.workspaceBusy(ctx, workspace.Folder); err != nil {
		failErr(w, err)
		return
	} else if busy != "" {
		fail(w, http.StatusConflict, busy+" is running in "+workspace.Folder+": change the home after the run ends")
		return
	}
	if err := os.RemoveAll(layout.HomeDir(h.opts.Root, workspace.Folder)); err != nil {
		failErr(w, err)
		return
	}
	output, err := h.makeHome(ctx, workspace.Folder, url, branch)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
		return
	}
	workspace.HomeURL, workspace.HomeBranch = url, branch
	if err := h.opts.Workspaces.Update(ctx, workspace); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace, "output": output})
}

// workspaceBusy names a job with an active run in a workspace, "" for
// none.
func (h *Handlers) workspaceBusy(ctx context.Context, name string) (string, error) {
	if h.opts.Runs == nil {
		return "", nil
	}
	active, err := h.opts.Runs.Active(ctx)
	if err != nil {
		return "", err
	}
	named := h.workspaceFolders(ctx)
	for _, run := range active {
		if workspace.Contains(name, named, run.Folder) {
			return run.Folder + "/" + run.Job, nil
		}
	}
	return "", nil
}

// deleteWorkspace removes the projects inside a workspace, their rows
// and their checkouts, their managed env files, its secrets and its keys
// with their grants, and then the workspace: its row and its directory,
// home and src with everything in them. The default workspace, the
// root, loses its row and the root's home and src, and the root
// directory stays with the named workspaces in it: the instance then
// runs without a default workspace. It refuses, a 409, while a run of
// the workspace is active or while the workspace holds a variable: a
// delete never takes a variable with it.
func (h *Handlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/delete") {
		return
	}
	ctx := r.Context()
	ws, err := h.opts.Workspaces.GetByFolder(ctx, param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	if busy, err := h.workspaceBusy(ctx, ws.Folder); err != nil {
		failErr(w, err)
		return
	} else if busy != "" {
		fail(w, http.StatusConflict, busy+" is running in "+ws.Folder+": delete the workspace after the run ends")
		return
	}
	if h.opts.Secrets != nil {
		secrets, err := h.opts.Secrets.Encrypted(ctx, ws.Folder)
		if err != nil {
			failErr(w, err)
			return
		}
		if len(secrets) > 0 {
			fail(w, http.StatusConflict, fmt.Sprintf("the workspace holds %d variables: delete them first, so none goes unseen", len(secrets)))
			return
		}
	}
	if err := h.emptyWorkspace(ctx, ws); err != nil {
		failErr(w, err)
		return
	}
	if err := h.opts.Workspaces.Delete(ctx, ws.ID); err != nil {
		failErr(w, err)
		return
	}
	dirs := []string{layout.WorkspaceDir(h.opts.Root, ws.Folder)}
	if ws.Folder == model.Default {
		dirs = []string{layout.HomeDir(h.opts.Root, ws.Folder), layout.SrcDir(h.opts.Root, ws.Folder)}
	}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			failErr(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// emptyWorkspace removes what a workspace holds in the database: the
// rows of the projects inside it, their managed env files, its secrets
// and its keys with their grants. The directories are the caller's.
func (h *Handlers) emptyWorkspace(ctx context.Context, ws *model.Workspace) error {
	projects, err := h.opts.Projects.List(ctx)
	if err != nil {
		return err
	}
	named := h.workspaceFolders(ctx)
	for _, project := range projects {
		if !workspace.Contains(ws.Folder, named, project.Folder) {
			continue
		}
		if err := h.opts.Projects.Delete(ctx, project.ID); err != nil {
			return err
		}
	}
	if h.opts.Secrets != nil {
		if _, err := h.opts.Secrets.Forget(ctx, ws.Folder); err != nil {
			return err
		}
	}
	if err := h.forgetEnvFiles(ctx, ws.Folder, named); err != nil {
		return err
	}
	if h.opts.Keys != nil {
		return h.opts.Keys.RemoveScope(ctx, keystore.ScopeWorkspace(ws.ID))
	}
	return nil
}

// promptWorkspace puts a prompt to Claude in a workspace, as POST
// /api/prompts does in a folder, and answers the same: the note, which
// the board answers by running @claude-prompt there as the user. Claude
// Code works in the workspace's home, TASK_UI_HOME.
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

// workspaceRows returns every workspace, none without the storage or on
// an error.
func (h *Handlers) workspaceRows(ctx context.Context) []model.Workspace {
	if h.opts.Workspaces == nil {
		return nil
	}
	rows, _ := h.opts.Workspaces.List(ctx)
	return rows
}

// dirOf returns the directory of a folder ID, as the layout places it
// among the named workspaces.
func (h *Handlers) dirOf(ctx context.Context, id string) string {
	return layout.Dir(h.opts.Root, h.workspaceFolders(ctx), id)
}

// forgetEnvFiles deletes the managed env files of a workspace's
// folders, its own and its projects': for internal, the root's folders outside
// the named workspaces.
func (h *Handlers) forgetEnvFiles(ctx context.Context, name string, named []string) error {
	if h.opts.Env == nil {
		return nil
	}
	files, err := h.opts.Env.Managed.All(ctx)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !workspace.Contains(name, named, file.Folder) {
			continue
		}
		if err := h.opts.Env.Managed.Delete(ctx, file.ID); err != nil {
			return err
		}
	}
	return nil
}
