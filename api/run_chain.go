package api

import (
	"context"
	"slices"
	"time"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// The kinds of chain a run is part of.
const (
	// ChainSummary is a Run and summarize: the job's run, then Claude's
	// summary of it once it starts. A job's run with no summary to come is
	// no chain.
	ChainSummary = "summary"
	// ChainWorkflow is a workflow run: the runs of its steps in order,
	// then the run of its prompt. A step that runs another workflow, or a
	// skipped one, has no run in it. Claude's runs are those of the
	// workflow run's prompt threads that no step names yet, oldest first: a
	// prompt step names its run once it ends, so a prompt step in progress
	// takes the first, and the workflow's prompt the last.
	ChainWorkflow = "workflow"
	// ChainBatch is a command on hosts: a run a host, oldest first, then
	// Claude's summary of them with Run and summarize. A batch of one run
	// without a summary is no chain.
	ChainBatch = "batch"
)

// RunChain is the chain of runs a run is part of, as the run window
// follows it: where the run stands in it, the runs beside it, and
// whether the chain has its last run.
type RunChain struct {
	// ID names the chain: the job's run of a summary, the workflow run,
	// the batch.
	ID   string `json:"id"`
	Kind string `json:"kind"`
	// Position is the run's place in the chain, from 1.
	Position int `json:"position"`
	// Total is how many runs the chain holds once complete, 0 while
	// that is not known.
	Total int `json:"total"`
	// Prev and Next are the runs before and after this one; Next is
	// empty until the next run starts.
	Prev string `json:"prev,omitempty"`
	Next string `json:"next,omitempty"`
	// Last is the newest run of the chain that started, unless that run
	// is this one: where a window opened on the chain goes.
	Last string `json:"last,omitempty"`
	// Complete says no more runs join the chain.
	Complete bool `json:"complete"`
}

// NewRunChain places current among the started runs of a chain, in order, or returns nil when they lack it.
// Total counts the runs to come as well; a complete chain's total is the runs it has.
func NewRunChain(kind, id string, runs []string, total int, complete bool, current string) *RunChain {
	at := slices.Index(runs, current)
	if at < 0 {
		return nil
	}
	c := &RunChain{ID: id, Kind: kind, Position: at + 1, Total: total, Complete: complete}
	if complete || c.Total < len(runs) {
		c.Total = len(runs)
	}
	if at > 0 {
		c.Prev = runs[at-1]
	}
	if at+1 < len(runs) {
		c.Next = runs[at+1]
		c.Last = runs[len(runs)-1]
	}
	return c
}

// SummaryChain is the chain of a Run and summarize, as ChainSummary describes it.
// Asked says a summary comes once the job ends.
func SummaryChain(current string, job model.Run, summary *model.Run, asked bool) *RunChain {
	runs := []string{job.ID}
	if summary != nil {
		runs = append(runs, summary.ID)
	} else if !asked {
		return nil
	}
	return NewRunChain(ChainSummary, job.ID, runs, 2, summary != nil && model.Finished(summary.State), current)
}

// BatchChain is the chain of a batch of host runs, as ChainBatch describes it.
// Asked says a summary comes once the batch ends.
func BatchChain(current, batchID string, batch []model.Run, summary *model.Run, asked bool) *RunChain {
	runs := make([]string, 0, len(batch)+1)
	complete := true
	for _, run := range batch {
		runs = append(runs, run.ID)
		complete = complete && model.Finished(run.State)
	}
	total := len(batch)
	switch {
	case summary != nil:
		runs = append(runs, summary.ID)
		total++
		complete = complete && model.Finished(summary.State)
	case asked:
		total++
		complete = false
	}
	if total < 2 {
		return nil
	}
	return NewRunChain(ChainBatch, batchID, runs, total, complete, current)
}

// WorkflowChain is the ChainWorkflow chain of a workflow run; prompt says the workflow has a prompt.
func WorkflowChain(current string, run model.WorkflowRun, steps []model.WorkflowRunStep, prompt bool, claude []string) *RunChain {
	var runs []string
	total := 0
	for _, step := range steps {
		if step.WorkflowRef != "" || step.State == model.StepSkipped {
			continue
		}
		total++
		switch {
		case step.RunID != "":
			runs = append(runs, step.RunID)
		case step.IsPrompt() && step.State == model.StateRunning && len(claude) > 0:
			runs = append(runs, claude[0])
			claude = claude[1:]
		}
	}
	if prompt {
		total++
		if len(claude) > 0 {
			runs = append(runs, claude[len(claude)-1])
		}
	}
	return NewRunChain(ChainWorkflow, run.ID, runs, total, model.Finished(run.State), current)
}

// chain returns the chain a run is part of, nil for none: a workflow
// run's step or prompt, else a summary's or a batch's.
func (h *Handlers) chain(ctx context.Context, run model.Run) *RunChain {
	if c := h.workflowChain(ctx, run); c != nil {
		return c
	}
	return h.summaryChain(ctx, run)
}

// workflowChain finds the workflow run that started a run, by its step
// or by its prompt's thread, and its chain.
func (h *Handlers) workflowChain(ctx context.Context, run model.Run) *RunChain {
	if h.opts.WorkflowRuns == nil {
		return nil
	}
	id := ""
	if step, err := h.opts.WorkflowRuns.StepOf(ctx, run.ID); err == nil {
		id = step.WorkflowRunID
	} else if thread := h.threadOf(ctx, run); thread != nil {
		id = thread.AboutWorkflowRunID
	}
	if id == "" {
		return nil
	}
	wr, err := h.opts.WorkflowRuns.Get(ctx, id)
	if err != nil {
		return nil
	}
	steps, err := h.opts.WorkflowRuns.Steps(ctx, id)
	if err != nil {
		return nil
	}
	prompt := false
	if h.opts.Workflows != nil {
		if wf, err := h.opts.Workflows.Get(ctx, wr.WorkflowID); err == nil {
			prompt = wf.HasPrompt()
		}
	}
	// Claude's runs: the threads about the workflow run, of its prompt
	// steps and its prompt, oldest first, that no step names.
	var claude []string
	if h.opts.Board != nil {
		threads, _ := h.opts.Board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: id})
		for i := len(threads) - 1; i >= 0; i-- {
			if threads[i].RunID != "" && !slices.ContainsFunc(steps, func(s model.WorkflowRunStep) bool { return s.RunID == threads[i].RunID }) {
				claude = append(claude, threads[i].RunID)
			}
		}
	}
	return WorkflowChain(run.ID, *wr, steps, prompt, claude)
}

// summaryChain finds the chain of a Run and summarize, or of a batch:
// from the job's run, or from Claude's run that summarizes it.
func (h *Handlers) summaryChain(ctx context.Context, run model.Run) *RunChain {
	base := &run
	if run.IsPrompt() {
		thread := h.threadOf(ctx, run)
		if !note.IsSummary(thread) {
			return nil
		}
		got, err := h.opts.Runs.Get(ctx, thread.AboutRunID)
		if err != nil {
			return nil
		}
		base = got
	}
	if base.BatchID != "" {
		batch, err := h.opts.Runs.List(ctx, storage.RunFilter{BatchID: base.BatchID, Limit: 200})
		if err != nil || len(batch) == 0 {
			return nil
		}
		slices.Reverse(batch)
		summary, asked := h.summaryOf(ctx, batch[0].ID)
		return BatchChain(run.ID, base.BatchID, batch, summary, asked)
	}
	summary, asked := h.summaryOf(ctx, base.ID)
	return SummaryChain(run.ID, *base, summary, asked)
}

// threadOf returns the thread one of Claude's runs answers, nil for
// another run.
func (h *Handlers) threadOf(ctx context.Context, run model.Run) *model.Note {
	if h.opts.Board == nil || !run.IsPrompt() {
		return nil
	}
	thread, err := h.opts.Board.Notes.ByRun(ctx, run.ID)
	if err != nil {
		return nil
	}
	return thread
}

// summaryRun returns the run of Claude's summary of a run, nil until it
// starts, and whether there is a summary to come or there: a summary
// thread that did not fail before its run.
func (h *Handlers) summaryRun(ctx context.Context, runID string) (*model.Run, bool) {
	if h.opts.Board == nil {
		return nil, false
	}
	threads, err := h.opts.Board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutRunID: runID})
	if err != nil {
		return nil, false
	}
	for i := len(threads) - 1; i >= 0; i-- {
		thread := &threads[i]
		if !note.IsSummary(thread) {
			continue
		}
		replies, _ := h.opts.Board.Notes.Replies(ctx, []string{thread.ID})
		id := note.SummaryRunID(thread, replies[thread.ID])
		if id == "" {
			return nil, thread.State != model.NoteFailed
		}
		if session, ok := h.opts.Runner.Get(id); ok {
			snapshot := session.Snapshot()
			return &snapshot, true
		}
		summary, err := h.opts.Runs.Get(ctx, id)
		if err != nil {
			return nil, false
		}
		return summary, true
	}
	return nil, false
}

// summaryOf returns the run of a run's summary and whether one comes:
// its thread exists, or Run and summarize waits on the run.
func (h *Handlers) summaryOf(ctx context.Context, runID string) (*model.Run, bool) {
	summary, found := h.summaryRun(ctx, runID)
	return summary, found || h.asking(runID)
}

// asking reports whether Run and summarize waits on the run, to start
// its summary once it ends.
func (h *Handlers) asking(runID string) bool {
	_, ok := h.summarizing.Load(runID)
	return ok
}

// How the attach socket follows a chain after its run ended: it looks
// for the next run every chainPoll, for chainWait at the most.
const (
	chainPoll = time.Second
	chainWait = 30 * time.Minute
)

// FollowChain looks up a chain until it has a next run or is complete,
// and sends it then, the window's cue to move on or to stop; it sends a
// chain that changed on the way, as a total that became known. It ends
// without a send when the run is no chain, the window went, or the wait
// is over.
func FollowChain(lookup func() *RunChain, send func(*RunChain) error, tick <-chan time.Time, gone <-chan struct{}, timeout <-chan time.Time) {
	var sent *RunChain
	for {
		c := lookup()
		if c == nil {
			return
		}
		if sent == nil || *c != *sent {
			if send(c) != nil {
				return
			}
			sent = c
		}
		if c.Next != "" || c.Complete {
			return
		}
		select {
		case <-tick:
		case <-gone:
			return
		case <-timeout:
			return
		}
	}
}
