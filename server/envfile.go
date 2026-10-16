package server

import (
	"log/slog"
	"net/http"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
)

// envFileOf resolves the folder and env file a request names. Only files
// a source declares can be edited.
func (svc *Server) envFileOf(r *http.Request, data *pageData) (*folder.Folder, string, bool) {
	f := svc.visibleFolder(r, data)
	if f == nil {
		return nil, "", false
	}
	path := param(r, "*")
	for _, declared := range f.EnvFiles {
		if declared == path {
			return f, path, true
		}
	}
	return f, path, false
}

// envFilePage renders the editor of a folder's env file.
func (svc *Server) envFilePage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	f, path, ok := svc.envFileOf(r, data)
	if !ok {
		svc.notFound(w, r, data, "No such env file. Only files a source reads can be edited.")
		return
	}
	file, err := svc.env.File(f, path)
	if err != nil {
		svc.notFound(w, r, data, err.Error())
		return
	}
	data.Title = path + " | " + f.Title
	data.Folder = f
	data.EnvFile = &file
	data.Secrets = svc.secrets.Scoped(f.ID)
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "envfile.tpl", data)
}

// envFileSave stores the content and regenerates the file in the folder.
func (svc *Server) envFileSave(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBody)
	data := svc.pageData(r, "")
	f, path, ok := svc.envFileOf(r, data)
	if !ok {
		http.NotFound(w, r)
		return
	}
	page := "/f/" + f.ID + "/env/" + path
	if err := svc.env.Save(f, envfile.File{Path: path, Content: r.PostFormValue("content")}); err != nil {
		redirectWith(w, r, page, "", err)
		return
	}
	slog.Info("env file saved", "folder", f.ID, "file", path, "by", currentUser(r).Name)
	redirectWith(w, r, page, "Saved "+path+" and generated it in the folder.", nil)
}
