package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Catalogue lists the folders and finds one. folder.Discovery is one; the
// composition builds a fresh one per call so configuration changes show.
type Catalogue interface {
	List() []*folder.Folder
	Find(id string) *folder.Folder
}

// visibleFolders returns the folders the user may see.
func (h *Handlers) visibleFolders(user *model.User) []*folder.Folder {
	var result []*folder.Folder
	for _, f := range h.opts.Catalogue.List() {
		if user.CanSee(f.ID) {
			result = append(result, f)
		}
	}
	return result
}

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
// see them. The second result says what is wrong when it cannot.
func (h *Handlers) target(r *http.Request, folderID, jobID string) (*folder.Folder, *folder.Job, string) {
	user := access.UserFrom(r.Context())
	f := h.visibleFolder(user, folderID)
	if f == nil {
		return nil, nil, "no such folder"
	}
	job := f.Job(jobID)
	if job == nil {
		return f, nil, "no such job"
	}
	return f, job, ""
}
