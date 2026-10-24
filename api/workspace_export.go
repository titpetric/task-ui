package api

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workspace"
)

// exportWorkspace answers with a workspace as a bundle another instance
// imports: its row, its projects and which is home, the grants of its SSH keys, its env
// files, its secrets in plain text and the SSH keys its runs get,
// encrypted to the passphrase, at least MinPassphrase characters. No
// instance key and no encryption key goes. Schedules, workflows, runs
// and artifacts stay. A POST keeps the passphrase out of the URL and
// the logs. Takes workspace/edit.
func (h *Handlers) exportWorkspace(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	var req WorkspaceExportRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := workspace.CheckPassphrase(req.Passphrase); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
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
	manifest.Workspace.HomeProjectID = ""

	projects, err := h.opts.Projects.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if !in(project.Folder) || project.Folder == name {
			continue
		}
		if project.ID == row.HomeProjectID {
			manifest.HomeProject = workspace.Relative(name, project.Folder)
		}
		project.ID, project.CreatedAt, project.UpdatedAt = "", nil, nil
		project.Folder = workspace.Relative(name, project.Folder)
		manifest.Projects = append(manifest.Projects, project)
	}
	var keyFiles map[string][]byte
	if h.opts.Keys != nil {
		var roles, users map[string][]string
		keyFiles, roles, users, err = h.opts.Keys.Bundle(ctx, keystore.ScopeWorkspace(row.ID))
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
		secrets, err := h.opts.Secrets.Plain(ctx, name)
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
	return &workspace.Bundle{Manifest: manifest, WorkFactor: h.opts.ExportWorkFactor, Keys: keyFiles}, nil
}
