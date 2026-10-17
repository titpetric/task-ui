package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/envfile"
)

// EnvRequest saves the managed content of an env file.
type EnvRequest struct {
	Content string `json:"content"`
}

// EnvView is an env file with its content, for admins.
type EnvView struct {
	envfile.File
	Content string `json:"content"`
	// Secrets are the names appended to the file when it is generated.
	Secrets []string `json:"secrets"`
}

// listEnv returns the env files a folder's sources read, without content.
func (h *Handlers) listEnv(w http.ResponseWriter, r *http.Request) {
	f := h.visibleFolder(access.UserFrom(r.Context()), param(r, "folder"))
	if f == nil {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": h.opts.Env.Files(r.Context(), f)})
}

// envView loads a file with what admins may see of it.
func (h *Handlers) envView(r *http.Request, f, path string) (*EnvView, error) {
	fold := h.opts.Catalogue.Find(f)
	if fold == nil {
		return nil, storageNotFound
	}
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

// getEnv returns an env file with its content. Admins only: the content
// may hold what the folder's own file keeps.
func (h *Handlers) getEnv(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	view, err := h.envView(r, param(r, "folder"), param(r, "*"))
	if err != nil {
		fail(w, statusOf(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// putEnv saves the managed content of an env file and generates it in
// the folder, with the secrets appended.
func (h *Handlers) putEnv(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req EnvRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	fold := h.opts.Catalogue.Find(param(r, "folder"))
	if fold == nil {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	file := envfile.File{Folder: fold.ID, Path: param(r, "*"), Content: req.Content}
	if err := h.opts.Env.Save(r.Context(), fold, file); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	view, err := h.envView(r, fold.ID, file.Path)
	if err != nil {
		fail(w, statusOf(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}
