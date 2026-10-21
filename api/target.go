package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// visibleFolder returns a folder the user may see, or nil.
func (h *Handlers) visibleFolder(user *model.User, id string) *folder.Folder {
	f := h.opts.Catalogue.Find(id)
	if f == nil || !user.CanSee(f.ID) {
		return nil
	}
	return f
}

// canRun reports whether the user may run a job of a folder.
func canRun(user *model.User, f *folder.Folder, job *folder.Job) bool {
	return f != nil && job != nil && user.CanRun(f.ID, f.Allow, job.Allow)
}

// target resolves the folder and job a request names, as the user may
// see them. The third result says what is wrong when it cannot.
func (h *Handlers) target(r *http.Request, folderID, jobID string) (*folder.Folder, *folder.Job, string) {
	f, job, _, problem := h.targetIn(r, folderID, "", jobID)
	return f, job, problem
}

// targetIn resolves a job of a directory inside a folder, as target
// does: dir is a slash path inside the folder, and the job one of the
// targets Folder.Sub reads there, or a skill atkins applies there. It
// returns the folder itself, whose rules and checkout the run takes, the
// job, and the directory the job runs in, its own Dir, "" for the
// folder's: a skill matched in a parent runs in that parent. A directory
// Path refuses, sealed, through .git or out by a symlink, is no such
// directory.
func (h *Handlers) targetIn(r *http.Request, folderID, dir, jobID string) (*folder.Folder, *folder.Job, string, string) {
	user := access.UserFrom(r.Context())
	f := h.visibleFolder(user, folderID)
	if f == nil {
		return nil, nil, "", "no such folder"
	}
	sub := f.Sub(dir)
	if sub == nil {
		return f, nil, "", "no such directory"
	}
	job := sub.Job(jobID)
	if job == nil {
		return f, nil, "", "no such job"
	}
	if job.Dir == "" {
		return f, job, "", ""
	}
	at, ok := f.Path(job.Dir)
	if !ok {
		return f, nil, "", "no such directory"
	}
	return f, job, at, ""
}

// requireAdmin answers anyone but an admin with 403 and reports it.
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}
