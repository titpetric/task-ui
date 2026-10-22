package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// workflowSummaryPrompt is what Claude reads a workflow run's report with.
const workflowSummaryPrompt = "The input is a markdown report of one run of a workflow in task-ui, a CI and task runner: the workflow, how it ended, and each step in order with the job it ran, how that ended and the end of its terminal output. " +
	claude.RunSummary

// workflowRunSummary returns the summary a workflow run has, from
// storage: 404 without one.
func (h *Handlers) workflowRunSummary(w http.ResponseWriter, r *http.Request) {
	run, ok := h.summaryWorkflowRun(w, r)
	if !ok {
		return
	}
	summary, err := h.opts.Summaries.Get(r.Context(), run.ID)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaryView(summary, true))
}

// summarizeWorkflowRun returns a finished workflow run's summary: the one
// it has, or one Claude writes now from the report of its steps, kept by
// the run's ID beside the runs' own.
func (h *Handlers) summarizeWorkflowRun(w http.ResponseWriter, r *http.Request) {
	run, ok := h.summaryWorkflowRun(w, r)
	if !ok {
		return
	}
	if kept, err := h.opts.Summaries.Get(r.Context(), run.ID); err == nil {
		writeJSON(w, http.StatusOK, summaryView(kept, true))
		return
	}
	if run.State == model.StateRunning {
		fail(w, http.StatusConflict, "the workflow is still running; summarize it after it ends")
		return
	}
	steps, err := h.opts.WorkflowRuns.Steps(r.Context(), run.ID)
	if err != nil {
		failErr(w, err)
		return
	}
	text, err := h.summarize(r, workflowSummaryPrompt, h.workflowReport(r, run, steps))
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

// summaryWorkflowRun returns the workflow run a summary route names, when
// the user may see every step of it and the API keeps summaries.
func (h *Handlers) summaryWorkflowRun(w http.ResponseWriter, r *http.Request) (*model.WorkflowRun, bool) {
	if h.opts.Summaries == nil {
		fail(w, http.StatusNotFound, "summaries aren't available on this instance")
		return nil, false
	}
	run, err := h.opts.WorkflowRuns.Get(r.Context(), param(r, "id"))
	if err != nil || !h.visibleRun(r, run.ID) {
		fail(w, http.StatusNotFound, storage.ErrNotFound.Error())
		return nil, false
	}
	return run, true
}

// workflowReport is the markdown Claude reads: the workflow run, then
// each step with its job, how it ended, and the end of its output, the
// lines shared out among the steps.
func (h *Handlers) workflowReport(r *http.Request, run *model.WorkflowRun, steps []model.WorkflowRunStep) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Workflow %s, run %s\n\n", run.Name, run.ID)
	fmt.Fprintf(&b, "Ended %s", run.State)
	if run.Message != "" {
		fmt.Fprintf(&b, " (%s)", run.Message)
	}
	fmt.Fprintf(&b, " after %s; started by %s.\n", run.Elapsed().Round(time.Millisecond), run.UserName)
	share := summaryLines
	if len(steps) > 0 {
		share = max(200, summaryLines/len(steps))
	}
	for i, step := range steps {
		name := step.Folder + " / " + step.Job
		if step.WorkflowRef != "" {
			name = "workflow " + step.WorkflowRef
		}
		fmt.Fprintf(&b, "\n## Step %d: %s, %s\n", i+1, name, step.State)
		if step.RunID == "" {
			continue
		}
		stepRun, err := h.opts.Runs.Get(r.Context(), step.RunID)
		if err != nil {
			continue
		}
		fmt.Fprintf(&b, "\nExit %d, after %s.\n", stepRun.ExitCode, stepRun.Elapsed().Round(time.Millisecond))
		output, err := h.runOutput(stepRun)
		if err != nil {
			continue
		}
		lines := strings.Split(output, "\n")
		if len(lines) > share {
			lines = lines[len(lines)-share:]
		}
		fmt.Fprintf(&b, "\n```\n%s\n```\n", strings.ReplaceAll(strings.Join(lines, "\n"), "```", "'''"))
	}
	return b.String()
}
