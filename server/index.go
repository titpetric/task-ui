package server

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
)

// index renders the overview: folders, active runs, recent history.
func (svc *Server) index(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Overview")
	if svc.config.History {
		for _, record := range svc.history.Recent(50) {
			if data.User.CanSee(record.Folder) && len(data.Recent) < 15 {
				data.Recent = append(data.Recent, record)
			}
		}
	}
	svc.render(w, r, "index.tpl", data)
}

// visibleFolder finds the folder the route names, when the user may see it.
func (svc *Server) visibleFolder(r *http.Request, data *pageData) *folder.Folder {
	id := param(r, "folder")
	for _, f := range data.Folders {
		if f.ID == id {
			return f
		}
	}
	return nil
}
