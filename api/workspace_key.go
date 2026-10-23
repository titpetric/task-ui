package api

import (
	"net/http"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// keysWorkspace returns the workspace a key route names, answering the
// request when there is none or the key folder is off. The route checks
// the permission first.
func (h *Handlers) keysWorkspace(w http.ResponseWriter, r *http.Request) *model.Workspace {
	if h.keysOff(w) {
		return nil
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return nil
	}
	return workspace
}

// listWorkspaceKeys returns a workspace's SSH keys, public halves and
// fingerprints only, with the roles and the users each goes to; a key
// with neither goes to every run in the workspace. A workspace has no
// encryption key: the instance's encrypts every secret. Takes
// workspace/view.
func (h *Handlers) listWorkspaceKeys(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	scope := keystore.ScopeWorkspace(workspace.ID)
	keys, err := h.opts.Keys.Of(r.Context(), scope, keystore.PurposeSSH)
	if err != nil {
		failErr(w, err)
		return
	}
	if keys == nil {
		keys = []keystore.Key{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": keys, "dir": h.opts.Keys.Dir})
}

// generateWorkspaceKey makes an SSH key of a workspace.
func (h *Handlers) generateWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (*keystore.Key, error) {
		return h.opts.Keys.Generate(r.Context(), h.workspaceSpec(workspace, req))
	})
}

// uploadWorkspaceKey stores a private key as an SSH key of a workspace.
// A key the folder holds, under any name or purpose, answers 409.
func (h *Handlers) uploadWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (*keystore.Key, error) {
		return h.opts.Keys.Import(r.Context(), h.workspaceSpec(workspace, req), []byte(req.Content))
	})
}

// workspaceSpec returns what a request makes for a workspace: its name
// under workspace_<folder>_, an SSH key of the workspace's scope.
func (h *Handlers) workspaceSpec(workspace *model.Workspace, req WorkspaceKeyRequest) keystore.Spec {
	kr := req.keyRequest()
	spec := kr.spec(keystore.ScopeWorkspace(workspace.ID))
	what := spec.Name
	if what == "" {
		what = "ssh"
	}
	spec.Name = keystore.WorkspaceKeyName(workspace.Folder, what)
	return spec
}

// addWorkspaceKey reads a key request, adds the SSH key, links it to
// the roles or users the request names, and answers with its public
// half. A request for an encryption key answers 400: one encryption
// key per instance, the instance's.
func (h *Handlers) addWorkspaceKey(w http.ResponseWriter, r *http.Request, add func(*model.Workspace, WorkspaceKeyRequest) (*keystore.Key, error)) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	var req WorkspaceKeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Encrypt {
		fail(w, http.StatusBadRequest, ErrWorkspaceEncrypt.Error())
		return
	}
	key, err := add(workspace, req)
	if err != nil {
		failKey(w, err)
		return
	}
	if err := h.link(r, key.Name, req.keyRequest()); err != nil {
		failKey(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key})
}

// workspaceSSHKey returns an SSH key of the workspace by name, answering
// the request when there is none.
func (h *Handlers) workspaceSSHKey(w http.ResponseWriter, r *http.Request, workspace *model.Workspace) *keystore.Key {
	key, err := h.opts.Keys.Get(r.Context(), param(r, "name"))
	if err != nil || key.Scope != keystore.ScopeWorkspace(workspace.ID) || key.Encrypts() {
		failErr(w, storage.ErrNotFound)
		return nil
	}
	return key
}

// updateWorkspaceKey sets the roles and the users whose runs get one of
// a workspace's SSH keys; with none of either, every run in the
// workspace gets it. Takes workspace/edit.
func (h *Handlers) updateWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	key := h.workspaceSSHKey(w, r, workspace)
	if key == nil {
		return
	}
	var req WorkspaceKeyRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Policies == nil {
		req.Policies = []string{}
	}
	if req.Users == nil {
		req.Users = []string{}
	}
	if err := h.opts.Keys.Link(r.Context(), key.Name, req.Policies, trimmed(req.Users)); err != nil {
		failKey(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": key.Name, "policies": req.Policies, "users": req.Users})
}

// deleteWorkspaceKey retires one of a workspace's SSH keys and drops
// its links. Takes workspace/edit.
func (h *Handlers) deleteWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	key := h.workspaceSSHKey(w, r, workspace)
	if key == nil {
		return
	}
	if err := h.opts.Keys.Remove(r.Context(), key.Name); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
