package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// keysWorkspace returns the workspace a key route names, for an admin,
// answering the request when there is none or the keyring is off.
func (h *Handlers) keysWorkspace(w http.ResponseWriter, r *http.Request) *model.Workspace {
	if !requireAdmin(w, r) {
		return nil
	}
	if h.opts.WorkspaceKeys == nil {
		fail(w, http.StatusNotFound, "workspace keys are off")
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
// fingerprints only: the encryption key, and the SSH keys with their
// backups and the roles and the users each goes to; a key with neither
// goes to every run in the workspace. Admins only.
func (h *Handlers) listWorkspaceKeys(w http.ResponseWriter, r *http.Request) {
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	roles, users := map[string][]string{}, map[string][]string{}
	if h.opts.KeyGrants != nil {
		var err error
		if roles, users, err = h.opts.KeyGrants.All(r.Context(), workspace.ID); err != nil {
			failErr(w, err)
			return
		}
	}
	keyring := h.opts.WorkspaceKeys
	writeJSON(w, http.StatusOK, map[string]any{
		"encrypt": keyring.EncryptKey(workspace.ID), "keys": keyring.SSH(workspace.ID).List(),
		"dir": keyring.Dir(workspace.ID), "policies": roles, "users": users,
	})
}

// generateWorkspaceKey makes a new SSH key pair for a workspace, or a
// new encryption key; the previous pair of the name stays as a backup.
func (h *Handlers) generateWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (sshkey.Key, error) {
		comment := strings.TrimSpace(req.Comment)
		if req.Encrypt {
			return h.opts.WorkspaceKeys.RotateEncrypt(workspace.ID, comment)
		}
		return h.opts.WorkspaceKeys.SSH(workspace.ID).GenerateAs(req.Type, strings.TrimSpace(req.Name), comment)
	})
}

// uploadWorkspaceKey stores a private key as a workspace's SSH key of
// its type, or as its encryption key.
func (h *Handlers) uploadWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	h.addWorkspaceKey(w, r, func(workspace *model.Workspace, req WorkspaceKeyRequest) (sshkey.Key, error) {
		comment := strings.TrimSpace(req.Comment)
		if req.Encrypt {
			return h.opts.WorkspaceKeys.UploadEncrypt(workspace.ID, []byte(req.Content), comment)
		}
		return h.opts.WorkspaceKeys.SSH(workspace.ID).UploadAs([]byte(req.Content), strings.TrimSpace(req.Name), comment)
	})
}

// addWorkspaceKey reads a key request, adds the key, sets the grants
// of an SSH key the request names, and answers with the key's public
// half.
func (h *Handlers) addWorkspaceKey(w http.ResponseWriter, r *http.Request, add func(*model.Workspace, WorkspaceKeyRequest) (sshkey.Key, error)) {
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
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if !req.Encrypt && (req.Policies != nil || req.Users != nil) {
		if err := h.grantWorkspaceKey(r, workspace, key.Name, req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key})
}

// grantWorkspaceKey sets the roles and the users whose runs get one of
// a workspace's SSH keys.
func (h *Handlers) grantWorkspaceKey(r *http.Request, workspace *model.Workspace, name string, req WorkspaceKeyRequest) error {
	if h.opts.KeyGrants == nil {
		return errors.New("workspace key grants are off")
	}
	users := make([]string, 0, len(req.Users))
	for _, user := range req.Users {
		users = append(users, strings.TrimSpace(user))
	}
	return h.opts.KeyGrants.Set(r.Context(), workspace.ID, name, req.Policies, users)
}

// workspaceSSHKey reports whether a workspace has an SSH key of the
// name, a backup among them when backups count.
func (h *Handlers) workspaceSSHKey(workspace *model.Workspace, name string, backups bool) bool {
	for _, key := range h.opts.WorkspaceKeys.SSH(workspace.ID).List() {
		if key.Name == name && (backups || !key.Backup) {
			return true
		}
	}
	return false
}

// updateWorkspaceKey sets the roles and the users whose runs get one of
// a workspace's SSH keys; with none of either, every run in the
// workspace gets it. Admins only.
func (h *Handlers) updateWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	name := param(r, "name")
	if !h.workspaceSSHKey(workspace, name, false) {
		failErr(w, storage.ErrNotFound)
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
	if err := h.grantWorkspaceKey(r, workspace, name, req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": name, "policies": req.Policies, "users": req.Users})
}

// deleteWorkspaceKey removes one of a workspace's SSH keys, or a backup
// of one, and its grants. The encryption key is not an SSH key: an
// admin rotates or replaces it, and no route deletes it. Admins only.
func (h *Handlers) deleteWorkspaceKey(w http.ResponseWriter, r *http.Request) {
	workspace := h.keysWorkspace(w, r)
	if workspace == nil {
		return
	}
	name := param(r, "name")
	if !h.workspaceSSHKey(workspace, name, true) {
		failErr(w, storage.ErrNotFound)
		return
	}
	if err := h.opts.WorkspaceKeys.SSH(workspace.ID).Delete(name); err != nil {
		failErr(w, err)
		return
	}
	if h.opts.KeyGrants != nil && !strings.Contains(name, ".bak-") {
		if err := h.opts.KeyGrants.Delete(r.Context(), workspace.ID, name); err != nil {
			failErr(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
