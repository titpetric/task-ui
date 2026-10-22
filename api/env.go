package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
)

// EnvRequest saves the managed content of an env file.
type EnvRequest struct {
	Content string `json:"content"`
}

// EnvView is an env file with its content, for admins.
type EnvView struct {
	envfile.File
	Content string `json:"content"`
	// Secrets are the names the generator appends to the file.
	Secrets []string `json:"secrets"`
}

// envTarget resolves the folder and the file a route names: the folder
// alone lists its files, folder/path is one of them.
func (h *Handlers) envTarget(r *http.Request) (*folder.Folder, string) {
	f, rest := folder.Resolve(h.opts.Catalogue, param(r, "*"))
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		return nil, ""
	}
	return f, rest
}

// listEnv returns the env files a folder's sources read, without content.
func (h *Handlers) listEnv(w http.ResponseWriter, r *http.Request, f *folder.Folder) {
	writeJSON(w, http.StatusOK, map[string]any{"files": h.opts.Env.Files(r.Context(), f)})
}

// envView loads a file with what admins may see of it.
func (h *Handlers) envView(r *http.Request, fold *folder.Folder, path string) (*EnvView, error) {
	file, err := h.opts.Env.File(r.Context(), fold, path)
	if err != nil {
		return nil, err
	}
	secrets, err := h.opts.Secrets.For(r.Context(), fold.ID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(secrets))
	for name := range secrets {
		names = append(names, name)
	}
	return &EnvView{File: file, Content: file.Content, Secrets: sorted(names)}, nil
}

// getEnv lists a folder's env files, or returns one with its content.
// The content takes env/view: it may hold what the folder's own
// file keeps.
func (h *Handlers) getEnv(w http.ResponseWriter, r *http.Request) {
	f, path := h.envTarget(r)
	if f == nil {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	if path == "" {
		h.listEnv(w, r, f)
		return
	}
	if !h.require(w, r, "env/view") {
		return
	}
	view, err := h.envView(r, f, path)
	if err != nil {
		fail(w, statusOf(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// putEnv saves the managed content of an env file and generates it in
// the folder, with the secrets appended. Content takes env/edit, and
// no content, which clears it, env/delete.
func (h *Handlers) putEnv(w http.ResponseWriter, r *http.Request) {
	if !h.can(r, "env/edit") && !h.require(w, r, "env/delete") {
		return
	}
	var req EnvRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	// Saving no content clears the file: a delete.
	if strings.TrimSpace(req.Content) == "" {
		if !h.require(w, r, "env/delete") {
			return
		}
	} else if !h.require(w, r, "env/edit") {
		return
	}
	fold, path := h.envTarget(r)
	if fold == nil || path == "" {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	file := envfile.File{Folder: fold.ID, Path: path, Content: req.Content}
	if err := h.opts.Env.Save(r.Context(), fold, file); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	view, err := h.envView(r, fold, file.Path)
	if err != nil {
		fail(w, statusOf(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}
