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

// listWorkspaceKeys returns a workspace's keys, public halves and
// fingerprints only: the encryption key, and the SSH keys with the
// roles and the users each goes to; a key with neither goes to every
// run in the workspace. Takes workspace/view.
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
	writeJSON(w, http.StatusOK, map[string]any{"encrypt": h.opts.Keys.EncryptKey(r.Context(), scope), "keys": keys, "dir": h.opts.Keys.Dir})
}

// generateWorkspaceKey makes an SSH key of a workspace, or its
// encryption key, new material for the one it has, for which it writes
// the workspace's secrets again.
func (h *Handlers) generateWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (*keystore.Key, error) {
		spec := h.workspaceSpec(workspace, req)
		if current := h.opts.Keys.EncryptKey(r.Context(), spec.Scope); req.Encrypt && current != nil {
			return h.opts.Keys.Rotate(r.Context(), current.Name)
		}
		return h.opts.Keys.Generate(r.Context(), spec)
	})
}

// uploadWorkspaceKey stores a private key as an SSH key of a workspace,
// or as its encryption key. A key the folder holds, under any name or
// purpose, answers 409.
func (h *Handlers) uploadWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (*keystore.Key, error) {
		return h.opts.Keys.Import(r.Context(), h.workspaceSpec(workspace, req), []byte(req.Content))
	})
}

// workspaceSpec returns what a request makes for a workspace: its name
// under workspace_<folder>_, the purpose, the workspace's scope.
func (h *Handlers) workspaceSpec(workspace *model.Workspace, req WorkspaceKeyRequest) keystore.Spec {
	kr := req.keyRequest()
	spec := kr.spec(keystore.ScopeWorkspace(workspace.ID))
	what := spec.Name
	if req.Encrypt {
		spec.Purpose, what = keystore.PurposeEncrypt, "encryption"
	}
	if what == "" {
		what = "ssh"
	}
	spec.Name = keystore.WorkspaceKeyName(workspace.Folder, what)
	return spec
}

// addWorkspaceKey reads a key request, adds the key, links an SSH key
// the request names roles or users for, and answers with the key's
// public half; a new encryption key writes the secrets again.
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
	key, err := add(workspace, req)
	if err != nil {
		failKey(w, err)
		return
	}
	if !req.Encrypt {
		if err := h.link(r, key.Name, req.keyRequest()); err != nil {
			failKey(w, err)
			return
		}
	}
	body := map[string]any{"key": key}
	if req.Encrypt && h.opts.Secrets != nil {
		h.rekeyWorkspace(r, workspace, body)
	}
	writeJSON(w, http.StatusCreated, body)
}

// rekeyWorkspace writes a workspace's secrets again for its encryption
// key and adds the count to the body. The previous material stays as a
// backup, so a rekey that fails is a field of the body, not a failure:
// the backup still opens the secrets.
func (h *Handlers) rekeyWorkspace(r *http.Request, workspace *model.Workspace, body map[string]any) {
	n, err := h.opts.Secrets.RekeyWorkspace(r.Context(), workspace.Folder)
	body["rekeyed"] = n
	if err != nil {
		body["rekey_error"] = err.Error()
	}
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
// its links. The encryption key is no SSH key: an admin rotates it to new
// material, and no route deletes it. Takes workspace/edit.
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
