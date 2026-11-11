package api_test

import (
	"errors"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A run's place in a chain: the runs beside it, and a total that a
// complete chain takes from the runs it holds.
func TestNewRunChain(t *testing.T) {
	t.Parallel()
	assert.Nil(t, api.NewRunChain(api.ChainWorkflow, "w1", []string{"a", "b"}, 3, false, "x"), "not in the chain")

	first := api.NewRunChain(api.ChainWorkflow, "w1", []string{"a", "b"}, 3, false, "a")
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 1, Total: 3, Next: "b", Last: "b"}, *first)
	mid := api.NewRunChain(api.ChainWorkflow, "w1", []string{"a", "b", "c"}, 0, false, "b")
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 2, Total: 3, Prev: "a", Next: "c", Last: "c"}, *mid)
	// Viewed from its first run, a chain that went on names its newest run.
	back := api.NewRunChain(api.ChainWorkflow, "w1", []string{"a", "b", "c"}, 3, true, "a")
	assert.Equal(t, "c", back.Last)
	done := api.NewRunChain(api.ChainWorkflow, "w1", []string{"a", "b"}, 5, true, "b")
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 2, Total: 2, Prev: "a", Complete: true}, *done)
}

// Run and summarize: the job's run, then the summary's, which the job's
// run knows of while it runs; a run without a summary is no chain.
func TestSummaryChain(t *testing.T) {
	t.Parallel()
	job := model.Run{ID: "r1", State: model.StateRunning}
	assert.Nil(t, api.SummaryChain("r1", job, nil, false), "no summary")

	before := api.SummaryChain("r1", job, nil, true)
	assert.Equal(t, api.RunChain{ID: "r1", Kind: api.ChainSummary, Position: 1, Total: 2}, *before)

	summary := &model.Run{ID: "r2", State: model.StateRunning}
	started := api.SummaryChain("r1", job, summary, false)
	assert.Equal(t, api.RunChain{ID: "r1", Kind: api.ChainSummary, Position: 1, Total: 2, Next: "r2", Last: "r2"}, *started)
	on := api.SummaryChain("r2", job, summary, false)
	assert.Equal(t, api.RunChain{ID: "r1", Kind: api.ChainSummary, Position: 2, Total: 2, Prev: "r1"}, *on)

	summary.State = model.StateOK
	end := api.SummaryChain("r2", job, summary, false)
	assert.True(t, end.Complete)
	assert.Equal(t, "", end.Next)
}

// A workflow run: its steps' runs in order, no run for a skipped step
// or one that runs another workflow, and its summary's run last.
func TestWorkflowChain(t *testing.T) {
	t.Parallel()
	running := model.WorkflowRun{ID: "w1", State: model.StateRunning}
	steps := []model.WorkflowRunStep{
		{Position: 0, RunID: "s1", State: model.StateOK},
		{Position: 1, Kind: model.StepWorkflow, WorkflowRef: "other", WorkflowRunRef: "w2", State: model.StateOK},
		{Position: 2, RunID: "s2", State: model.StateRunning},
		{Position: 3, State: model.StepPending},
		{Position: 4, Kind: model.StepSummary, Folder: "_", Job: model.PromptJob, State: model.StepPending},
	}
	mid := api.WorkflowChain("s2", running, steps, nil)
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 2, Total: 4, Prev: "s1"}, *mid)
	first := api.WorkflowChain("s1", running, steps, nil)
	assert.Equal(t, "s2", first.Next)

	steps[2].State, steps[3].RunID, steps[3].State = model.StateOK, "s3", model.StateOK
	last := api.WorkflowChain("s3", running, steps, nil)
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 3, Total: 4, Prev: "s2"}, *last)
	assert.False(t, last.Complete, "the summary is to come")

	// The summary in progress stands in for Claude's run, then names it.
	steps[4].State = model.StateRunning
	asking := api.WorkflowChain("p1", running, steps, []string{"p1"})
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 4, Total: 4, Prev: "s3"}, *asking)
	assert.Equal(t, "p1", api.WorkflowChain("s3", running, steps, []string{"p1"}).Next)
	done := running
	done.State = model.StateOK
	steps[4].State, steps[4].RunID = model.StateOK, "p1"
	summary := api.WorkflowChain("p1", done, steps, nil)
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 4, Total: 4, Prev: "s3", Complete: true}, *summary)

	// A step that failed skips the rest: the chain ends with it.
	steps[3] = model.WorkflowRunStep{Position: 3, State: model.StepSkipped}
	steps[4] = model.WorkflowRunStep{Position: 4, Kind: model.StepSummary, State: model.StepSkipped}
	failed := running
	failed.State = model.StateFailed
	end := api.WorkflowChain("s2", failed, steps, nil)
	assert.Equal(t, api.RunChain{ID: "w1", Kind: api.ChainWorkflow, Position: 2, Total: 2, Prev: "s1", Complete: true}, *end)
	assert.Nil(t, api.WorkflowChain("x", running, steps, nil))
}

// A batch of host runs, then their summary with Run and summarize.
func TestBatchChain(t *testing.T) {
	t.Parallel()
	batch := []model.Run{{ID: "h1", State: model.StateOK}, {ID: "h2", State: model.StateRunning}}
	plain := api.BatchChain("h1", "b1", batch, nil, false)
	assert.Equal(t, api.RunChain{ID: "b1", Kind: api.ChainBatch, Position: 1, Total: 2, Next: "h2", Last: "h2"}, *plain)

	asked := api.BatchChain("h2", "b1", batch, nil, true)
	assert.Equal(t, api.RunChain{ID: "b1", Kind: api.ChainBatch, Position: 2, Total: 3, Prev: "h1"}, *asked)

	batch[1].State = model.StateOK
	summary := &model.Run{ID: "p1", State: model.StateOK}
	end := api.BatchChain("p1", "b1", batch, summary, false)
	assert.Equal(t, api.RunChain{ID: "b1", Kind: api.ChainBatch, Position: 3, Total: 3, Prev: "h2", Complete: true}, *end)

	assert.Nil(t, api.BatchChain("h1", "b1", batch[:1], nil, false), "one run is no chain")
}

// The socket sends the chain as it stands, then again once the next run
// started, and stops there.
func TestFollowChain(t *testing.T) {
	t.Parallel()
	looks := []*api.RunChain{
		{ID: "r1", Kind: api.ChainSummary, Position: 1, Total: 2},
		{ID: "r1", Kind: api.ChainSummary, Position: 1, Total: 2},
		{ID: "r1", Kind: api.ChainSummary, Position: 1, Total: 2, Next: "r2"},
	}
	at := 0
	lookup := func() *api.RunChain {
		c := looks[at]
		at++
		return c
	}
	var sent []api.RunChain
	send := func(c *api.RunChain) error {
		sent = append(sent, *c)
		return nil
	}
	tick := make(chan time.Time, 3)
	for range 3 {
		tick <- time.Time{}
	}
	api.FollowChain(lookup, send, tick, nil, nil)
	assert.Len(t, sent, 2, "unchanged chains go once")
	assert.Equal(t, "r2", sent[1].Next)
	assert.Equal(t, 3, at)

	// No chain sends nothing; a gone window and a failed send stop it.
	sent = nil
	api.FollowChain(func() *api.RunChain { return nil }, send, tick, nil, nil)
	assert.Len(t, sent, 0)
	gone := make(chan struct{})
	close(gone)
	waiting := func() *api.RunChain { return &api.RunChain{ID: "r1", Position: 1} }
	api.FollowChain(waiting, send, nil, gone, nil)
	assert.Len(t, sent, 1)
	timeout := make(chan time.Time, 1)
	timeout <- time.Time{}
	api.FollowChain(waiting, send, nil, nil, timeout)
	assert.Len(t, sent, 2)
	calls := 0
	api.FollowChain(waiting, func(*api.RunChain) error { calls++; return errors.New("closed") }, tick, nil, nil)
	assert.Equal(t, 1, calls)
}
