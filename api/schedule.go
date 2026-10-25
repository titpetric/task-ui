package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

// ScheduleRequest creates or changes a schedule: of a job, Folder and
// Job, or of a workflow, Workflow.
type ScheduleRequest struct {
	Folder     string `json:"folder"`
	Job        string `json:"job"`
	Expression string `json:"expression"`
	Enabled    *bool  `json:"enabled"`
	// Workflow is the ID of the workflow the schedule starts, in place
	// of a folder and a job: a run of every step as the schedule's user,
	// with no input. Such a schedule takes no args and no triage.
	Workflow string `json:"workflow,omitempty"`
	// User is who the schedule runs as, by name or ID: an admin picks
	// anyone, everyone else only themselves. Empty is the caller on
	// create, and no change on update.
	User string `json:"user"`
	// Args follow the job's argv after --, for a job that takes args:
	// the command of @host:<host>, which needs one. On update, empty
	// keeps them.
	Args RunArgs `json:"args,omitempty"`
	// Triage has Claude triage a run of the schedule that fails: a
	// thread on the run's page and a note on the board, as the
	// schedule's user, who needs prompt/edit. Nil keeps it on update.
	Triage *bool `json:"triage"`
}

// errWorkflowSchedule refuses the fields of a job's schedule on a
// workflow's.
const errWorkflowSchedule = "a workflow schedule takes no folder, job, args or triage"

// onOff is the row's value of a switch of a request, 1 or 0.
func onOff(on bool) int64 {
	if on {
		return 1
	}
	return 0
}

// jobArgs checks the args of a schedule or a workflow step: a command
// for @host:<host>, which needs one, and for any other job only when it
// takes args. It returns what is wrong, "" for nothing.
func jobArgs(job *folder.Job, args []string) string {
	switch {
	case job.IsHost() && strings.TrimSpace(strings.Join(args, " ")) == "":
		return job.ID + " needs a command as its args"
	case len(args) > 0 && !job.TakesArgs:
		return runner.ErrNoArgs.Error()
	}
	return ""
}

// seesSchedule reports whether the user of a request may see a
// schedule: its folder, or every step of its workflow.
func (h *Handlers) seesSchedule(r *http.Request, schedule *model.Schedule) bool {
	if schedule.IsWorkflow() {
		_, _, err := h.workflowByID(r, schedule.WorkflowID)
		return err == nil
	}
	return access.UserFrom(r.Context()).CanSee(schedule.Folder)
}

// listSchedules returns the schedules of the folders the user may see,
// and of the workflows whose every step they see.
func (h *Handlers) listSchedules(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/view") {
		return
	}
	schedules, err := h.opts.Schedules.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	visible := []model.Schedule{}
	for _, schedule := range schedules {
		if h.seesSchedule(r, &schedule) {
			visible = append(visible, schedule)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedules": visible})
}

// createSchedule adds a schedule for a job the user may run, or for a
// workflow whose every step they may run.
func (h *Handlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/edit") {
		return
	}
	var req ScheduleRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Workflow != "" {
		h.createWorkflowSchedule(w, r, req)
		return
	}
	user := access.UserFrom(r.Context())
	if req.Job == model.HostBare {
		fail(w, http.StatusBadRequest, folder.ErrHostBare.Error())
		return
	}
	f, job, problem := h.target(r, req.Folder, req.Job)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if job.IsShell() {
		fail(w, http.StatusBadRequest, folder.ErrShellUnattended.Error())
		return
	}
	if problem := jobArgs(job, req.Args); problem != "" {
		fail(w, http.StatusBadRequest, problem)
		return
	}
	if !h.canRun(r, f, job) {
		fail(w, http.StatusForbidden, "you can view this job, but you don't have permission to schedule it")
		return
	}
	runAs, status, problem := h.runAs(r, req.User, f.ID, job.ID)
	if problem != "" {
		fail(w, status, problem)
		return
	}
	if runAs == nil {
		runAs = user
	}
	next, err := scheduler.NextRun(req.Expression, time.Now())
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	schedule := &model.Schedule{Folder: f.ID, Job: job.ID, UserID: runAs.ID, UserName: runAs.Name, Expression: req.Expression, Enabled: 1, NextRunAt: &next, Args: model.EncodeList(req.Args)}
	if req.Enabled != nil && !*req.Enabled {
		schedule.Enabled = 0
	}
	if req.Triage != nil {
		schedule.Triage = onOff(*req.Triage)
	}
	if err := h.opts.Schedules.Create(r.Context(), schedule); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

// createWorkflowSchedule adds a schedule of a workflow, for a user who
// may run every step of it, as startWorkflow asks: no folder, job, args
// or triage with it (400), a workflow the user sees (404).
func (h *Handlers) createWorkflowSchedule(w http.ResponseWriter, r *http.Request, req ScheduleRequest) {
	if req.Folder != "" || req.Job != "" || len(req.Args) > 0 || (req.Triage != nil && *req.Triage) {
		fail(w, http.StatusBadRequest, errWorkflowSchedule)
		return
	}
	wf, steps, err := h.workflowByID(r, req.Workflow)
	if err != nil {
		failErr(w, err)
		return
	}
	if step := h.mayRun(r, steps); step != "" {
		fail(w, http.StatusForbidden, "you don't have permission to run "+step)
		return
	}
	runAs, status, problem := h.runAsWorkflow(r, req.User, steps)
	if problem != "" {
		fail(w, status, problem)
		return
	}
	if runAs == nil {
		runAs = access.UserFrom(r.Context())
	}
	next, err := scheduler.NextRun(req.Expression, time.Now())
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	schedule := &model.Schedule{WorkflowID: wf.ID, UserID: runAs.ID, UserName: runAs.Name, Expression: req.Expression, Enabled: 1, NextRunAt: &next}
	if req.Enabled != nil && !*req.Enabled {
		schedule.Enabled = 0
	}
	if err := h.opts.Schedules.Create(r.Context(), schedule); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

// pickedUser resolves the user a schedule request names to run as, by
// name or ID. It returns nil for no name, and the caller for their own.
// Only an admin names someone else (403); a user nobody holds is a 400.
func (h *Handlers) pickedUser(r *http.Request, name string) (*model.User, int, string) {
	if name == "" {
		return nil, 0, ""
	}
	caller := access.UserFrom(r.Context())
	if name == caller.Name || name == caller.ID {
		return caller, 0, ""
	}
	if !caller.CanAdmin() {
		return nil, http.StatusForbidden, "only an admin picks who a schedule runs as"
	}
	user, err := h.opts.Users.GetByName(r.Context(), name)
	if err != nil {
		if user, err = h.opts.Users.Get(r.Context(), name); err != nil {
			return nil, http.StatusBadRequest, "no user " + name
		}
	}
	return user, 0, ""
}

// runAs resolves the user a schedule request names to run as, as
// pickedUser does, and checks that someone else may run the job: they
// see its folder and the run permission and the allow lists let them,
// a 400 otherwise.
func (h *Handlers) runAs(r *http.Request, name, folderID, jobID string) (*model.User, int, string) {
	user, status, problem := h.pickedUser(r, name)
	if user == nil || user == access.UserFrom(r.Context()) {
		return user, status, problem
	}
	as := r.WithContext(access.WithUser(r.Context(), user))
	f, job, problem := h.target(as, folderID, jobID)
	if problem != "" || job.IsShell() || !h.canRun(as, f, job) {
		return nil, http.StatusBadRequest, user.Name + " may not run " + folderID + "/" + jobID
	}
	return user, 0, ""
}

// runAsWorkflow resolves the user a workflow schedule request names to
// run as, as pickedUser does, and checks that someone else may run
// every step, a 400 naming the step otherwise.
func (h *Handlers) runAsWorkflow(r *http.Request, name string, steps []model.WorkflowStep) (*model.User, int, string) {
	user, status, problem := h.pickedUser(r, name)
	if user == nil || user == access.UserFrom(r.Context()) {
		return user, status, problem
	}
	as := r.WithContext(access.WithUser(r.Context(), user))
	if step := h.mayRun(as, steps); step != "" {
		return nil, http.StatusBadRequest, user.Name + " may not run " + step
	}
	return user, 0, ""
}

// scheduleOf returns the schedule a route names, when the user may see it.
func (h *Handlers) scheduleOf(r *http.Request) (*model.Schedule, error) {
	schedule, err := h.opts.Schedules.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, err
	}
	if !h.seesSchedule(r, schedule) {
		return nil, storage.ErrNotFound
	}
	return schedule, nil
}

// getSchedule returns one schedule.
func (h *Handlers) getSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/view") {
		return
	}
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

// mayChange reports whether the user may change a schedule: its owner,
// or anyone who may run its job, or every step of its workflow.
func (h *Handlers) mayChange(r *http.Request, schedule *model.Schedule) bool {
	user := access.UserFrom(r.Context())
	if user.ID == schedule.UserID && schedule.UserID != "" {
		return true
	}
	if schedule.IsWorkflow() {
		_, steps, err := h.workflowByID(r, schedule.WorkflowID)
		return err == nil && h.mayRun(r, steps) == ""
	}
	f, job, _ := h.target(r, schedule.Folder, schedule.Job)
	return h.canRun(r, f, job)
}

// updateSchedule changes the expression, the user, the args or the
// triage of a schedule, or switches it. A workflow's schedule takes no
// args or triage.
func (h *Handlers) updateSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/edit") {
		return
	}
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.mayChange(r, schedule) {
		fail(w, http.StatusForbidden, "only its owner, or someone who can run its job, can change this schedule")
		return
	}
	var req ScheduleRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if schedule.IsWorkflow() {
		if req.Folder != "" || req.Job != "" || len(req.Args) > 0 || req.Triage != nil {
			fail(w, http.StatusBadRequest, errWorkflowSchedule)
			return
		}
		_, steps, _ := h.workflowByID(r, schedule.WorkflowID)
		runAs, status, problem := h.runAsWorkflow(r, req.User, steps)
		if problem != "" {
			fail(w, status, problem)
			return
		}
		if runAs != nil {
			schedule.UserID, schedule.UserName = runAs.ID, runAs.Name
		}
	} else {
		runAs, status, problem := h.runAs(r, req.User, schedule.Folder, schedule.Job)
		if problem != "" {
			fail(w, status, problem)
			return
		}
		if runAs != nil {
			schedule.UserID, schedule.UserName = runAs.ID, runAs.Name
		}
		if len(req.Args) > 0 {
			_, job, _ := h.target(r, schedule.Folder, schedule.Job)
			if job == nil {
				fail(w, http.StatusNotFound, "no such job")
				return
			}
			if problem := jobArgs(job, req.Args); problem != "" {
				fail(w, http.StatusBadRequest, problem)
				return
			}
			schedule.Args = model.EncodeList(req.Args)
		}
		if req.Triage != nil {
			schedule.Triage = onOff(*req.Triage)
		}
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
		schedule.Enabled = onOff(*req.Enabled)
	}
	if err := h.opts.Schedules.Update(r.Context(), schedule); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

// deleteSchedule removes a schedule.
func (h *Handlers) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/delete") {
		return
	}
	schedule, err := h.scheduleOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.mayChange(r, schedule) {
		fail(w, http.StatusForbidden, "only its owner, or someone who can run its job, can change this schedule")
		return
	}
	if err := h.opts.Schedules.Delete(r.Context(), schedule.ID); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
