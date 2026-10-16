package server

import (
	"net/http"
)

// folderPage renders a folder with its jobs, by source.
func (svc *Server) folderPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	f := svc.visibleFolder(r, data)
	if f == nil {
		svc.notFound(w, r, data, "No such folder.")
		return
	}
	data.Title = f.Title
	data.Folder = f
	data.EnvFiles = svc.env.Files(f)
	data.Git = svc.git.Info(f.Dir)
	svc.render(w, r, "folder.tpl", data)
}
