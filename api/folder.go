package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/storage"
)

// listFolders returns the folders the user may see with their jobs, each
// with whether the user may run it and its latest run. ?folder= keeps
// one folder.
func (h *Handlers) listFolders(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	only := r.URL.Query().Get("folder")
	views := []FolderView{}
	for _, f := range h.opts.Catalogue.List() {
		if !user.CanSee(f.ID) || only != "" && f.ID != only {
			continue
		}
		view := FolderView{ID: f.ID, Title: f.Title, Description: f.Description, Error: f.Error, Errors: f.Errors, Jobs: []JobView{}}
		for _, job := range f.Jobs {
			last, err := h.opts.Runs.Last(r.Context(), storage.RunFilter{Folder: f.ID, Job: job.ID})
			if err != nil && !errors.Is(err, storage.ErrNotFound) {
				failErr(w, err)
				return
			}
			view.Jobs = append(view.Jobs, JobView{Job: job, CanRun: h.canRun(r, f, job), LastRun: last})
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, map[string]any{"folders": views})
}

// folderActions answers GET /folders/{id}/actions?dir= with the actions
// bar of the folder, or of dir inside it. Anyone who sees the folder
// reads it; can_run says who may run each job.
func (h *Handlers) folderActions(w http.ResponseWriter, r *http.Request) {
	id, ok := strings.CutSuffix(folderParam(r), "/actions")
	if !ok {
		fail(w, http.StatusNotFound, "no such route")
		return
	}
	user := access.UserFrom(r.Context())
	f := h.visibleFolder(user, id)
	if f == nil {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	dir := strings.Trim(r.URL.Query().Get("dir"), "/")
	place := f.Sub(dir)
	if place == nil {
		fail(w, http.StatusNotFound, "no such directory")
		return
	}
	view := ActionsView{Folder: f.ID, Dir: place.Rel, Skills: []SkillView{}, Jobs: []JobView{}}
	if h.opts.Git != nil && h.opts.Root != "" {
		view.Git = h.opts.Git.Status(place.Dir, h.opts.Root)
	}
	jobView := func(job *folder.Job) (JobView, error) {
		last, err := h.opts.Runs.Last(r.Context(), storage.RunFilter{Folder: f.ID, Dir: job.Dir, Job: job.ID})
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return JobView{}, err
		}
		return JobView{Job: job, CanRun: h.canRun(r, f, job), LastRun: last}, nil
	}
	for _, skill := range place.Skills() {
		sv := SkillView{Skill: skill, Jobs: []JobView{}}
		for _, job := range skill.Jobs {
			jv, err := jobView(job)
			if err != nil {
				failErr(w, err)
				return
			}
			sv.Jobs = append(sv.Jobs, jv)
		}
		view.Skills = append(view.Skills, sv)
	}
	for _, job := range place.Jobs {
		jv, err := jobView(job)
		if err != nil {
			failErr(w, err)
			return
		}
		view.Jobs = append(view.Jobs, jv)
	}
	writeJSON(w, http.StatusOK, view)
}
