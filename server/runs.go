package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/titpetric/task-ui/run"
)

// apiRuns lists the active runs the user can see.
func (svc *Server) apiRuns(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	runs := []*run.Run{}
	for _, active := range svc.runs.List() {
		if user.CanSee(active.Folder) {
			runs = append(runs, active)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(runs)
}

// apiStop kills an active run. Anyone allowed to run the job may stop it,
// which is the same permission Run force needs.
func (svc *Server) apiStop(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	f := svc.visibleFolder(r, data)
	job := f.Job(param(r, "job"))
	if job == nil {
		http.NotFound(w, r)
		return
	}
	if !data.User.CanRun(f, job) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if active, ok := svc.runs.Get(job.Key()); ok {
		slog.Info("run killed", "folder", f.ID, "job", job.ID, "by", data.User.Name, "owner", active.User)
		active.Kill()
	}
	if r.Header.Get("Accept") == "application/json" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, safeBack(r.FormValue("back")), http.StatusSeeOther)
}
