package api

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// RunRequest is what starts a run.
type RunRequest struct {
	Folder string `json:"folder"`
	// Dir is a directory inside the folder, a slash path: the job is one
	// of its own targets and runs there. Empty is the folder itself.
	Dir   string `json:"dir"`
	Job   string `json:"job"`
	Force bool   `json:"force"`
	// Args follow the job's argv after --, for a job that takes args.
	Args RunArgs `json:"args"`
	// Reason says why the user forces the run out of hours. Required then.
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
	switch {
	case run.IsPrompt():
		v.CanRun = h.endsPrompt(r, &run)
		return v
	case run.IsNote():
		v.CanRun = h.endsNote(r, &run)
		return v
	}
	f, job, _, _ := h.targetIn(r, run.Folder, run.Dir, run.Job)
	v.CanRun = h.canRun(r, f, job)
	return v
}

// endsPrompt reports whether the user of a request may end a prompt's
// run: the asker, or an admin, with prompt/edit.
func (h *Handlers) endsPrompt(r *http.Request, run *model.Run) bool {
	user := access.UserFrom(r.Context())
	return user != nil && (run.UserID == user.ID || user.CanAdmin()) && h.can(r, "prompt/edit")
}

// endsNote reports whether the user of a request may end the Notes
// agent's run: its starter, the agent's user, or an admin, with
// note/edit.
func (h *Handlers) endsNote(r *http.Request, run *model.Run) bool {
	user := access.UserFrom(r.Context())
	return user != nil && (run.UserID == user.ID || user.CanAdmin()) && h.can(r, "note/edit")
}

// listRuns returns the runs the user may see, newest first.
func (h *Handlers) listRuns(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	filter := storage.RunFilter{Folder: q.Get("folder"), Folders: user.FolderList(), Job: q.Get("job"), Dir: q.Get("dir"), State: q.Get("state"), Limit: queryInt(r, "limit", 50)}
	runs, err := h.opts.Runs.List(r.Context(), filter)
	if err != nil {
		failErr(w, err)
		return
	}
	views := []RunView{}
	prompts, notes := h.can(r, "prompt/view"), h.can(r, "note/view")
	for _, run := range runs {
		if user.SeesRun(&run, prompts, notes) {
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
	switch req.Job {
	case model.PromptJob:
		fail(w, http.StatusBadRequest, "a prompt starts from /api/prompts, with its text")
		return
	case model.NoteJob:
		fail(w, http.StatusBadRequest, "the Notes agent starts from the Notes board, on a note that names it")
		return
	}
	f, job, dir, problem := h.targetIn(r, req.Folder, req.Dir, req.Job)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if job.IsShell() && !h.can(r, "shell") {
		fail(w, http.StatusForbidden, "an interactive shell, or Claude in one, is for admins")
		return
	}
	if !h.canRun(r, f, job) {
		fail(w, http.StatusForbidden, "your role may view this job, but not run it")
		return
	}
	if len(req.Args) > 0 && !job.TakesArgs {
		fail(w, http.StatusBadRequest, runner.ErrNoArgs.Error())
		return
	}
	session, err := h.opts.Runner.Start(r.Context(), runner.Request{Folder: f, Job: job, User: user, Force: req.Force, Reason: req.Reason, Cols: req.Cols, Rows: req.Rows, Dir: dir, Args: req.Args})
	if errors.Is(err, runner.ErrBusy) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "already running", "run": h.view(r, session.Snapshot())})
		return
	}
	if errors.Is(err, runner.ErrOutOfHours) || errors.Is(err, runner.ErrReasonNeeded) {
		// The window prompts for the reason and forces; it needs the hours
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

// runOf returns the run a route names, when the user may see it: a run
// of a folder the user sees, a prompt's run there with prompt/view, the
// Notes agent's with note/view, or such a run the user started.
func (h *Handlers) runOf(r *http.Request) (*model.Run, error) {
	run, err := h.opts.Runs.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, err
	}
	if !access.UserFrom(r.Context()).SeesRun(run, h.can(r, "prompt/view"), h.can(r, "note/view")) {
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

// endRun stops or kills a run, for a user allowed to run its job, or
// to end a prompt's run.
func (h *Handlers) endRun(w http.ResponseWriter, r *http.Request, kill bool) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.view(r, *run).CanRun {
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

// frameOf returns the current frame of a run's session. With ?all=1 it
// is the run's whole output so far instead: its recording replayed into
// a tall terminal.
func (h *Handlers) frameOf(r *http.Request) (runner.Frame, error) {
	run, err := h.runOf(r)
	if err != nil {
		return runner.Frame{}, err
	}
	frame, _, _, err := h.replayOf(r, run)
	return frame, err
}

// replayOf returns the frame of a run as frameOf does, the lines of its
// whole output, and whether the last screen already shows them all: a
// run that ended and never scrolled. A run in progress takes its frame
// from its session, its lines not counted (0).
func (h *Handlers) replayOf(r *http.Request, run *model.Run) (runner.Frame, int, bool, error) {
	all := r.URL.Query().Get("all") != ""
	if session, ok := h.opts.Runner.Get(run.ID); ok && !all && !session.Status().Finished() {
		return session.Frame(), 0, false, nil
	}
	// The run ended or the request names ?all: the recording
	// holds both.
	if h.opts.Recordings == nil {
		return runner.Frame{}, 0, false, runner.ErrNotFound
	}
	cast, err := h.opts.Recordings.Read(run)
	if err != nil {
		if session, ok := h.opts.Runner.Get(run.ID); ok && !all {
			return session.Frame(), 0, false, nil
		}
		return runner.Frame{}, 0, false, runner.ErrNotFound
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return runner.Frame{}, 0, false, err
	}
	if all {
		return replay.Output, replay.Lines(), replay.Complete(), nil
	}
	return replay.Screen, replay.Lines(), replay.Complete(), nil
}

// frameSVG renders the last frame of a run as an image. It drops the empty
// rows at the bottom, so a tile shows what there is; ?full=1
// keeps the whole screen, ?all=1 renders the whole output. The
// X-Output-Lines header counts the lines of the whole output, and
// X-Output-Complete says the screen already shows them all.
func (h *Handlers) frameSVG(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	frame, lines, complete, err := h.replayOf(r, run)
	if err != nil {
		failErr(w, err)
		return
	}
	if lines > 0 {
		w.Header().Set("X-Output-Lines", strconv.Itoa(lines))
	}
	if complete {
		w.Header().Set("X-Output-Complete", "true")
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
