package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/note"
)

// report shows the report of a run, for a user who sees the run and
// has prompt/view: Claude's answer, or its summary of the run, as a
// document, without the terminal.
func (h *Handlers) report(w http.ResponseWriter, r *http.Request) {
	row, err := h.opts.Runs.Get(r.Context(), param(r, "id"))
	if err != nil || !h.runSeer(r)(row) {
		h.missing(w, r, "No run with that ID.")
		return
	}
	run := h.runView(*row)
	p := h.page(r, run.Name()+" report")
	var report *note.Report
	if h.opts.Board != nil && p.Can("prompt/view") {
		report = h.opts.Board.Report(r.Context(), row)
	}
	if report == nil {
		h.missing(w, r, "The run has no report: no summary, nor an answer of Claude's.")
		return
	}
	p.Run, p.Report = &run, report
	p.Crumbs = view.Crumbs("Runs", "/runs", run.Name(), "/runs/"+run.ID, "Report")
	about := run
	if report.AboutRunID != "" && report.AboutRunID != run.ID {
		if got, err := h.opts.Runs.Get(r.Context(), report.AboutRunID); err == nil && h.runSeer(r)(got) {
			about = h.runView(*got)
		}
	}
	p.ReportOf = &about
	h.render(w, http.StatusOK, "report", p)
}
