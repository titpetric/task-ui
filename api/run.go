package api

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// RunRequest is what starts a run.
type RunRequest struct {
	Folder string `json:"folder"`
	Job    string `json:"job"`
	Force  bool   `json:"force"`
	// Reason says why the run is forced out of hours. Required then.
	Reason string `json:"reason"`
	Cols   int    `json:"cols"`
	Rows   int    `json:"rows"`
}

// RunView is a run as the API shows it: the row, and the live status
// and the windows attached while the session is in memory.
type RunView struct {
	model.Run
	Live     bool           `json:"live"`
	Attached int            `json:"attached"`
	Status   *runner.Status `json:"status,omitempty"`
	// CanRun says whether the user asking may run the job, and so end
	// the run and type into it.
	CanRun bool `json:"can_run"`
}

// view decorates a run row with what the runner knows, for the user
// asking.
func (h *Handlers) view(r *http.Request, run model.Run) RunView {
	v := RunView{Run: run}
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		status := session.Status()
		v.Live, v.Attached, v.Status = true, session.Attached(), &status
		v.Run = session.Snapshot()
	}
	f, job, _ := h.target(r, run.Folder, run.Job)
	v.CanRun = canRun(access.UserFrom(r.Context()), f, job)
	return v
}

// listRuns returns the runs the user may see, newest first.
func (h *Handlers) listRuns(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	filter := storage.RunFilter{Folder: q.Get("folder"), Job: q.Get("job"), State: q.Get("state"), Limit: queryInt(r, "limit", 50)}
	runs, err := h.opts.Runs.List(r.Context(), filter)
	if err != nil {
		failErr(w, err)
		return
	}
	views := []RunView{}
	for _, run := range runs {
		if user.CanSee(run.Folder) {
			views = append(views, h.view(r, run))
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": views})
}

// startRun starts a job. A job that is running answers 409 with the run
// in the way, unless the request forces.
func (h *Handlers) startRun(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
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
		fail(w, http.StatusForbidden, "your role may view this job, but not run it")
		return
	}
	session, err := h.opts.Runner.Start(r.Context(), runner.Request{Folder: f, Job: job, User: user, Force: req.Force, Reason: req.Reason, Cols: req.Cols, Rows: req.Rows})
	if errors.Is(err, runner.ErrBusy) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "already running", "run": h.view(r, session.Snapshot())})
		return
	}
	if errors.Is(err, runner.ErrOutOfHours) || errors.Is(err, runner.ErrReasonNeeded) {
		// The window asks for the reason and forces; it needs the hours
		// to say what they are.
		writeJSON(w, statusOf(err), map[string]any{"error": err.Error(), "out_of_hours": true, "hours": f.HoursOf(job)})
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, h.view(r, session.Snapshot()))
}

// runOf returns the run a route names, when the user may see it.
func (h *Handlers) runOf(r *http.Request) (*model.Run, error) {
	run, err := h.opts.Runs.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, err
	}
	if !access.UserFrom(r.Context()).CanSee(run.Folder) {
		return nil, storage.ErrNotFound
	}
	return run, nil
}

// getRun returns one run.
func (h *Handlers) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.view(r, *run))
}

// endRun stops or kills a run, for a user allowed to run its job.
func (h *Handlers) endRun(w http.ResponseWriter, r *http.Request, kill bool) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	f, job, problem := h.target(r, run.Folder, run.Job)
	if problem != "" || !canRun(access.UserFrom(r.Context()), f, job) {
		fail(w, http.StatusForbidden, "your role may view this run, but not end it")
		return
	}
	if kill {
		err = h.opts.Runner.Kill(run.ID)
	} else {
		err = h.opts.Runner.Stop(run.ID)
	}
	if err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) stopRun(w http.ResponseWriter, r *http.Request) { h.endRun(w, r, false) }

func (h *Handlers) killRun(w http.ResponseWriter, r *http.Request) { h.endRun(w, r, true) }

// frameOf returns the current frame of a run's session.
func (h *Handlers) frameOf(r *http.Request) (runner.Frame, error) {
	run, err := h.runOf(r)
	if err != nil {
		return runner.Frame{}, err
	}
	session, ok := h.opts.Runner.Get(run.ID)
	if !ok {
		return runner.Frame{}, runner.ErrNotFound
	}
	return session.Frame(), nil
}

// frameSVG renders the last frame of a run as an image. The empty rows
// at the bottom are left out, so a tile shows what there is; ?full=1
// keeps the whole screen.
func (h *Handlers) frameSVG(w http.ResponseWriter, r *http.Request) {
	frame, err := h.frameOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if r.URL.Query().Get("full") == "" {
		frame = frame.Trim(8)
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(frame.SVG(runner.DarkPalette)))
}

// frameText renders the last frame of a run as text.
func (h *Handlers) frameText(w http.ResponseWriter, r *http.Request) {
	frame, err := h.frameOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(frame.Text()))
}

// cast serves the recording of a run.
func (h *Handlers) cast(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if run.CastPath == "" || h.opts.Recordings == nil {
		fail(w, http.StatusNotFound, "no recording")
		return
	}
	w.Header().Set("Content-Type", "application/x-asciicast")
	http.ServeFile(w, r, filepath.Join(h.opts.Recordings.Root, filepath.FromSlash(run.CastPath)))
}
