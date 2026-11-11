package api

import (
	"net/http"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/access"
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
	// A prompt's run repeats as the question it answered, not as a job:
	// POST /api/runs refuses the job a prompt runs under.
	if run.IsPrompt() || run.IsNote() {
		h.askAgain(w, r, run)
		return
	}
	h.start(w, r, RerunRequest(run))
}

// askAgain puts the question a prompt's run answered to the board again, in its own thread and as the caller, and answers what POST /api/prompts answers; 404 once the thread is gone.
// It takes what POST /api/prompts takes, because it starts the same run: prompt/edit, and no run's own token, which would be a builder starting another builder.
func (h *Handlers) askAgain(w http.ResponseWriter, r *http.Request, run *model.Run) {
	if !h.require(w, r, "prompt/edit") {
		return
	}
	if access.FromRun(r.Context()) {
		fail(w, http.StatusConflict, "a run's token can't start a prompt: do the work with the task tools")
		return
	}
	if h.opts.Board == nil {
		fail(w, http.StatusNotFound, "not found")
		return
	}
	note, err := h.opts.Board.Notes.ByRun(r.Context(), run.ID)
	if err != nil {
		fail(w, http.StatusNotFound, "the question this run answered is gone")
		return
	}
	h.ask(w, r, PromptRequest{Prompt: note.Body, Parent: note.ID, Cols: int(run.TermCols), Rows: int(run.TermRows)}, "")
}

// RerunRequest returns the request that runs a stored run again: its
// folder, directory, job and arguments, the job naming its host.
func RerunRequest(run *model.Run) RunRequest {
	return RunRequest{Folder: run.Folder, Dir: run.Dir, Job: run.Job, Args: run.ArgList()}
}
