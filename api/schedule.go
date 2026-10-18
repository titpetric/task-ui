package api

import (
	"net/http"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

// ScheduleRequest creates or changes a schedule.
type ScheduleRequest struct {
	Folder     string `json:"folder"`
	Job        string `json:"job"`
	Expression string `json:"expression"`
	Enabled    *bool  `json:"enabled"`
}

// listSchedules returns the schedules of the folders the user may see.
func (h *Handlers) listSchedules(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	schedules, err := h.opts.Schedules.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	visible := []model.Schedule{}
	for _, schedule := range schedules {
		if user.CanSee(schedule.Folder) {
			visible = append(visible, schedule)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedules": visible})
}

// createSchedule adds a schedule for a job the user may run.
func (h *Handlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	user := access.UserFrom(r.Context())
	f, job, problem := h.target(r, req.Folder, req.Job)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if !canRun(user, f, job) {
		fail(w, http.StatusForbidden, "your role may view this job, but not schedule it")
		return
	}
	next, err := scheduler.NextRun(req.Expression, time.Now())
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	schedule := &model.Schedule{Folder: f.ID, Job: job.ID, UserID: user.ID, UserName: user.Name, Expression: req.Expression, Enabled: 1, NextRunAt: &next}
	if req.Enabled != nil && !*req.Enabled {
		schedule.Enabled = 0
	}
	if err := h.opts.Schedules.Create(r.Context(), schedule); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

// scheduleOf returns the schedule a route names, when the user may see it.
func (h *Handlers) scheduleOf(r *http.Request) (*model.Schedule, error) {
	schedule, err := h.opts.Schedules.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, err
	}
	if !access.UserFrom(r.Context()).CanSee(schedule.Folder) {
		return nil, storage.ErrNotFound
	}
	return schedule, nil
}

// getSchedule returns one schedule.
func (h *Handlers) getSchedule(w http.ResponseWriter, r *http.Request) {
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

// mayChange reports whether the user may change a schedule: its owner,
// or anyone who may run its job.
func (h *Handlers) mayChange(r *http.Request, schedule *model.Schedule) bool {
	user := access.UserFrom(r.Context())
	if user.ID == schedule.UserID && schedule.UserID != "" {
		return true
	}
	f, job, _ := h.target(r, schedule.Folder, schedule.Job)
	return canRun(user, f, job)
}

// updateSchedule changes the expression or switches a schedule.
func (h *Handlers) updateSchedule(w http.ResponseWriter, r *http.Request) {
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.mayChange(r, schedule) {
		fail(w, http.StatusForbidden, "not your schedule")
		return
	}
	var req ScheduleRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Expression != "" {
		next, err := scheduler.NextRun(req.Expression, time.Now())
		if err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		schedule.Expression, schedule.NextRunAt = req.Expression, &next
	}
	if req.Enabled != nil {
		schedule.Enabled = 0
		if *req.Enabled {
			schedule.Enabled = 1
		}
	}
	if err := h.opts.Schedules.Update(r.Context(), schedule); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

// deleteSchedule removes a schedule.
func (h *Handlers) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.mayChange(r, schedule) {
		fail(w, http.StatusForbidden, "not your schedule")
		return
	}
	if err := h.opts.Schedules.Delete(r.Context(), schedule.ID); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
