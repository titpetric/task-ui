package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/markdown"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// summaryLines bounds the output a run summary reads: the end of it.
const summaryLines = 2000

// runSummaryPrompt is what Claude reads a run's output with.
const runSummaryPrompt = "The input is a markdown report of one run of a job in task-ui, a CI and task runner: the job, how it ended, and the end of its terminal output. " +
	claude.RunSummary

// runSummary returns the summary a run has, from storage: 404 without one.
func (h *Handlers) runSummary(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if h.opts.Summaries == nil {
		fail(w, http.StatusNotFound, "no summaries here")
		return
	}
	summary, err := h.opts.Summaries.Get(r.Context(), run.ID)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaryView(summary, true))
}

// summarizeRun returns a finished run's summary: the one it has, or one
// Claude writes from its output now, kept for the next ask.
func (h *Handlers) summarizeRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.runOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if h.opts.Summaries == nil {
		fail(w, http.StatusNotFound, "no summaries here")
		return
	}
	if kept, err := h.opts.Summaries.Get(r.Context(), run.ID); err == nil {
		writeJSON(w, http.StatusOK, summaryView(kept, true))
		return
	}
	if run.State == model.StateRunning {
		fail(w, http.StatusConflict, "the run is still going; summarize it when it ends")
		return
	}
	output, err := h.runOutput(run)
	if err != nil {
		fail(w, http.StatusNotFound, "the run left no output to summarize")
		return
	}
	text, err := h.summarize(r, runSummaryPrompt, runReport(run, output))
	if err != nil {
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	summary := &model.RunSummary{RunID: run.ID, Summary: text, UserName: access.UserFrom(r.Context()).Name}
	if err := h.opts.Summaries.Set(r.Context(), summary); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaryView(summary, false))
}

// runOutput is the end of a run's output, as text, from its recording.
func (h *Handlers) runOutput(run *model.Run) (string, error) {
	if h.opts.Recordings == nil {
		return "", runner.ErrNotFound
	}
	cast, err := h.opts.Recordings.Read(run)
	if err != nil {
		return "", err
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(replay.Output.Text(), " \n"), "\n")
	if len(lines) > summaryLines {
		lines = lines[len(lines)-summaryLines:]
	}
	text := strings.Join(lines, "\n")
	if strings.TrimSpace(text) == "" {
		return "", errors.New("no output")
	}
	return text, nil
}

// runReport is the markdown Claude reads: the run, how it ended, and its
// output in a fence.
func runReport(run *model.Run, output string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s, run %s\n\n", run.Name(), run.ID)
	fmt.Fprintf(&b, "Ended %s, exit %d, after %s; started by %s.\n", run.State, run.ExitCode, run.Elapsed().Round(time.Millisecond), run.UserName)
	fmt.Fprintf(&b, "\n```\n%s\n```\n", strings.ReplaceAll(output, "```", "'''"))
	return b.String()
}

// summaryView is a summary as the API returns it: the markdown, and the
// same as HTML for a page to show.
func summaryView(summary *model.RunSummary, kept bool) map[string]any {
	return map[string]any{
		"run_id":     summary.RunID,
		"summary":    summary.Summary,
		"html":       markdown.HTML([]byte(summary.Summary), "", ""),
		"user_name":  summary.UserName,
		"created_at": summary.CreatedAt,
		"kept":       kept,
	}
}
