package api

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workspace"
)

// exportWorkspace answers with a workspace as a bundle another instance
// imports: its row, its projects, the grants of its SSH keys, its env
// files, its secrets as ciphertext and its key files, encrypted to the
// passphrase. Schedules, workflows, runs and artifacts stay. A POST
// keeps the passphrase out of the URL. Takes workspace/edit.
func (h *Handlers) exportWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	var req WorkspaceExportRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Passphrase == "" {
		fail(w, http.StatusBadRequest, "a bundle needs a passphrase")
		return
	}
	row, err := h.opts.Workspaces.GetByFolder(r.Context(), param(r, "folder"))
	if err != nil {
		failErr(w, err)
		return
	}
	bundle, err := h.bundle(r.Context(), row)
	if err != nil {
		failErr(w, err)
		return
	}
	var out bytes.Buffer
	if err := bundle.Write(&out, req.Passphrase); err != nil {
		failErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(row.Folder+workspace.Extension))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(out.Bytes())
}

// bundle gathers what an export carries, every folder relative to the
// workspace.
func (h *Handlers) bundle(ctx context.Context, row *model.Workspace) (*workspace.Bundle, error) {
	named := h.workspaceFolders(ctx)
	name := row.Folder
	in := func(folder string) bool { return workspace.Contains(name, named, folder) }
	manifest := workspace.Manifest{Format: workspace.Format, Version: h.opts.Version, Workspace: *row}
	manifest.Workspace.ID, manifest.Workspace.CreatedAt, manifest.Workspace.UpdatedAt = "", nil, nil

	projects, err := h.opts.Projects.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if !in(project.Folder) || project.Folder == name {
			continue
		}
		project.ID, project.CreatedAt, project.UpdatedAt = "", nil, nil
		project.Folder = workspace.Relative(name, project.Folder)
		manifest.Projects = append(manifest.Projects, project)
	}
	if h.opts.KeyGrants != nil {
		roles, users, err := h.opts.KeyGrants.All(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		for key, list := range roles {
			for _, role := range list {
				manifest.Grants = append(manifest.Grants, model.WorkspaceKeyGrant{KeyName: key, Role: role})
			}
		}
		for key, list := range users {
			for _, user := range list {
				manifest.Grants = append(manifest.Grants, model.WorkspaceKeyGrant{KeyName: key, UserName: user})
			}
		}
	}
	if h.opts.Env != nil {
		files, err := h.opts.Env.Managed.All(ctx)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if !in(file.Folder) {
				continue
			}
			file.ID, file.CreatedAt, file.UpdatedAt = "", nil, nil
			file.Folder = workspace.Relative(name, file.Folder)
			manifest.EnvFiles = append(manifest.EnvFiles, file)
		}
	}
	if h.opts.Secrets != nil && h.opts.Secrets.Workspaces != nil {
		secrets, err := h.opts.Secrets.Encrypted(ctx, name)
		if err != nil {
			return nil, err
		}
		for _, secret := range secrets {
			if model.IsReserved(secret.Scope) {
				continue
			}
			secret.ID, secret.CreatedAt, secret.UpdatedAt = "", nil, nil
			secret.Scope = workspace.RelativeScope(name, secret.Scope)
			manifest.Secrets = append(manifest.Secrets, secret)
		}
	}
	bundle := &workspace.Bundle{Manifest: manifest, WorkFactor: h.opts.ExportWorkFactor}
	if h.opts.WorkspaceKeys != nil {
		if bundle.Keys, err = h.opts.WorkspaceKeys.Files(row.ID); err != nil {
			return nil, err
		}
	}
	return bundle, nil
}
