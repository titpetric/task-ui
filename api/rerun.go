package api

import (
	"net/http"

	"github.com/titpetric/task-ui/model"
)

// rerunRun runs a stored run again with its own inputs, as the user
// asking: 404 for a run the user may not see, 409 for one still going,
// and else what POST /api/runs answers.
func (h *Handlers) rerunRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	// A live session knows how the run stands; the row lags behind it.
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		*run = session.Snapshot()
	}
	if !run.Finished() {
		fail(w, http.StatusConflict, "you can view this run, but it hasn't ended yet")
		return
	}
	h.start(w, r, RerunRequest(run))
}

// RerunRequest returns the request that runs a stored run again: its
// folder, directory, job and arguments, the job naming its host.
func RerunRequest(run *model.Run) RunRequest {
	return RunRequest{Folder: run.Folder, Dir: run.Dir, Job: run.Job, Args: run.ArgList()}
}
