package server

import (
	"net/http"
	"path"

	"github.com/titpetric/task-ui/history"
)

// jobPage renders a job: its description, run actions and history.
func (svc *Server) jobPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	f := svc.visibleFolder(r, data)
	job := f.Job(param(r, "job"))
	if job == nil {
		svc.notFound(w, r, data, "No such job.")
		return
	}
	data.Title = job.Name + " | " + f.Title
	data.Folder = f
	data.Job = job
	if svc.config.History {
		data.History = svc.history.List(f.ID, job.ID)
	}
	svc.render(w, r, "job.tpl", data)
}

// historyFile serves a recording.
func (svc *Server) historyFile(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	f := svc.visibleFolder(r, data)
	job := f.Job(param(r, "job"))
	if job == nil || !svc.config.History {
		http.NotFound(w, r)
		return
	}
	file, err := svc.history.CastPath(history.Record{Folder: f.ID, Job: job.ID, ID: param(r, "id")})
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/x-asciicast")
	w.Header().Set("Content-Disposition", "inline; filename="+path.Base(file))
	http.ServeFile(w, r, file)
}
