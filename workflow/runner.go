// Package workflow runs workflows: named chains of jobs from any folders,
// started one after another as ordinary runs. A step starts when the one
// before it ended ok; the first that does not end ok ends the workflow
// run, and the runner skips the steps after it. A step names a job or
// another workflow; a workflow runs as a run of its own, and ends the
// step as it ended. The runner skips a disabled step and runs the rest.
// A workflow with a prompt puts it to Claude once its steps end, with
// their output, and its run ends with Claude's reply.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// ErrBusy reports a workflow that already has a run in progress.
var ErrBusy = errors.New("the workflow is running")

// ErrNotRunning refuses a stop of a workflow run that ended.
var ErrNotRunning = errors.New("the workflow run is not running")

// ErrNothing refuses a run of a workflow with no enabled step and no
// prompt.
var ErrNothing = errors.New("the workflow has no enabled step and no prompt")

// Catalogue finds the folder a step names. folder.Discovery is one.
type Catalogue interface {
	Find(id string) *folder.Folder
}

// Step is a workflow step resolved to its folder and job, or to the
// workflow it runs in place of one.
type Step struct {
	Folder *folder.Folder
	Job    *folder.Job
	// Workflow is the workflow the step runs, nil for a job.
	Workflow *model.Workflow
	// Disabled is a step the run skips, resolved to nothing.
	Disabled bool
}

// Prompter puts a workflow's prompt to Claude once a run's steps ended,
// as the user who started the run, with the run's input, and returns
// when Claude replied, nil, or the prompt failed; a stop of ctx ends
// the prompt's run. note.Board is one.
type Prompter interface {
	AskWorkflow(ctx context.Context, user *model.User, run *model.WorkflowRun, workflow *model.Workflow, input string) error
}

// Runner starts workflow runs and walks their steps.
type Runner struct {
	workflows *storage.WorkflowStorage
	runs      *storage.WorkflowRunStorage
	catalogue Catalogue
	manager   *runner.Manager
	// Events gets a workflow's topic told about its runs: `run` when
	// one starts, `step` as each step changes, `done` when it ended.
	// Nil tells nobody.
	Events *event.Bus
	// Prompter runs the prompt of a workflow that has one; nil fails a
	// run of such a workflow once its steps end.
	Prompter Prompter

	mu     sync.Mutex
	byFlow map[string]*active // workflow ID -> its run in progress
	byRun  map[string]*active // workflow run ID -> the same
}

// active is a workflow run in progress, of the workflow, with the
// run's input.
type active struct {
	run      *model.WorkflowRun
	workflow *model.Workflow
	input    string
	cancel   context.CancelFunc
	done     chan struct{}
}

// New returns a runner.
func New(workflows *storage.WorkflowStorage, runs *storage.WorkflowRunStorage, catalogue Catalogue, manager *runner.Manager) *Runner {
	return &Runner{
		workflows: workflows,
		runs:      runs,
		catalogue: catalogue,
		manager:   manager,
		byFlow:    map[string]*active{},
		byRun:     map[string]*active{},
	}
}

// Resolve finds the folder and job, or the workflow, of every enabled
// step, or says which one is missing.
func (r *Runner) Resolve(ctx context.Context, steps []model.WorkflowStep) ([]Step, error) {
	resolved := make([]Step, 0, len(steps))
	for _, step := range steps {
		if step.Disabled {
			resolved = append(resolved, Step{Disabled: true})
			continue
		}
		if step.IsWorkflow() {
			wf, err := r.workflows.Get(ctx, step.WorkflowRef)
			if err != nil {
				return nil, fmt.Errorf("no workflow %s", step.WorkflowRef)
			}
			resolved = append(resolved, Step{Workflow: wf})
			continue
		}
		f := r.catalogue.Find(step.Folder)
		var job *folder.Job
		if f != nil {
			job = f.Job(step.Job)
		}
		if job == nil {
			return nil, fmt.Errorf("no job %s / %s", step.Folder, step.Job)
		}
		if job.IsShell() {
			return nil, fmt.Errorf("%s / %s: %w", step.Folder, step.Job, folder.ErrShellUnattended)
		}
		resolved = append(resolved, Step{Folder: f, Job: job})
	}
	return resolved, nil
}

// Start begins a run of the workflow as the user, and walks its steps in
// the background; input goes to the workflow's prompt, after it. The
// caller checks the user may run every step, and the prompt.
func (r *Runner) Start(ctx context.Context, workflow *model.Workflow, user *model.User, input string) (*model.WorkflowRun, error) {
	steps, err := r.workflows.Steps(ctx, workflow.ID)
	if err != nil {
		return nil, err
	}
	if !workflow.Runs(steps) {
		return nil, ErrNothing
	}
	resolved, err := r.Resolve(ctx, steps)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, busy := r.byFlow[workflow.ID]; busy {
		return nil, ErrBusy
	}
	run := &model.WorkflowRun{WorkflowID: workflow.ID, Name: workflow.Name, UserID: user.ID, UserName: user.Name}
	if err := r.runs.Create(ctx, run, steps); err != nil {
		return nil, err
	}
	rows, err := r.runs.Steps(ctx, run.ID)
	if err != nil {
		return nil, err
	}

	walkCtx, cancel := context.WithCancel(context.Background())
	a := &active{run: run, workflow: workflow, input: input, cancel: cancel, done: make(chan struct{})}
	r.byFlow[workflow.ID], r.byRun[run.ID] = a, a
	snapshot := *run
	r.publish(workflow.ID, "run", RunEvent{RunID: run.ID, WorkflowID: workflow.ID, Name: run.Name, State: run.State, UserName: run.UserName})
	go r.walk(walkCtx, a, resolved, rows, user)
	return &snapshot, nil
}

// RunEvent is the payload of a workflow's `run` and `done` events.
type RunEvent struct {
	RunID      string `json:"run_id"`
	WorkflowID string `json:"workflow_id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Message    string `json:"message,omitempty"`
	UserName   string `json:"user_name"`
}

// StepEvent is the payload of a workflow's `step` event: the step as
// its row stands.
type StepEvent struct {
	RunID          string `json:"run_id"`
	Position       int64  `json:"position"`
	Folder         string `json:"folder,omitempty"`
	Job            string `json:"job,omitempty"`
	WorkflowRef    string `json:"workflow_ref,omitempty"`
	State          string `json:"state"`
	StepRunID      string `json:"step_run_id,omitempty"`
	WorkflowRunRef string `json:"workflow_run_ref,omitempty"`
}

// publish tells the workflow's topic, when there is a bus.
func (r *Runner) publish(workflowID, name string, data any) {
	if r.Events != nil {
		r.Events.Publish("workflow:"+workflowID, name, data)
	}
}

// Active returns the ID of the workflow's run in progress.
func (r *Runner) Active(workflowID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.byFlow[workflowID]; ok {
		return a.run.ID, true
	}
	return "", false
}

// Done returns a channel that closes when the workflow run ended; for a
// run not in progress it returns a closed channel.
func (r *Runner) Done(runID string) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.byRun[runID]; ok {
		return a.done
	}
	closed := make(chan struct{})
	close(closed)
	return closed
}

// Stop ends a workflow run: it stops the running step and skips the
// rest.
func (r *Runner) Stop(runID string) error {
	r.mu.Lock()
	a, ok := r.byRun[runID]
	r.mu.Unlock()
	if !ok {
		return ErrNotRunning
	}
	a.cancel()
	return nil
}

// walk runs the steps in order and records how each, and the whole,
// ended.
func (r *Runner) walk(ctx context.Context, a *active, steps []Step, rows []model.WorkflowRunStep, user *model.User) {
	defer func() {
		r.mu.Lock()
		delete(r.byFlow, a.run.WorkflowID)
		delete(r.byRun, a.run.ID)
		r.mu.Unlock()
		close(a.done)
	}()
	// The records outlive a stopped walk, so the runner writes them without
	// its context.
	store := context.Background()
	run := a.run
	run.State = model.StateOK

	for i, step := range steps {
		row := &rows[i]
		if run.State != model.StateOK || step.Disabled {
			row.State = model.StepSkipped
			r.saveStep(store, row)
			continue
		}
		row.State = model.StateRunning
		r.saveStep(store, row)
		var state, message string
		if step.Workflow != nil {
			state, row.WorkflowRunRef, message = r.nested(ctx, step.Workflow, user)
		} else {
			state, row.RunID, message = r.step(ctx, step, user)
		}
		row.State = state
		r.saveStep(store, row)
		switch {
		case state == model.StateOK:
		case state == model.StateStopped && ctx.Err() != nil:
			run.State, run.Message = model.StateStopped, "stopped at "+label(step)
		default:
			run.State, run.Message = model.StateFailed, fmt.Sprintf("%s %s", label(step), firstOf(message, state))
		}
	}
	if a.workflow.HasPrompt() && run.State != model.StateStopped {
		r.prompt(ctx, a, user)
	}
	if err := r.runs.Update(store, run); err != nil {
		slog.Warn("workflow run not saved", "workflow", run.Name, "err", err)
	}
	slog.Info("workflow finished", "workflow", run.Name, "user", run.UserName, "state", run.State)
	r.publish(run.WorkflowID, "done", RunEvent{RunID: run.ID, WorkflowID: run.WorkflowID, Name: run.Name, State: run.State, Message: run.Message, UserName: run.UserName})
}

// prompt runs the workflow's prompt once the steps ended, through the
// Prompter, and returns when Claude replied. The run keeps the steps'
// state, or fails with a prompt that fails after steps that ended ok; a
// stop during the prompt stops the run.
func (r *Runner) prompt(ctx context.Context, a *active, user *model.User) {
	run := a.run
	var err error
	if r.Prompter == nil {
		err = errors.New("prompts are off")
	} else {
		err = r.Prompter.AskWorkflow(ctx, user, run, a.workflow, a.input)
	}
	switch {
	case err == nil:
	case ctx.Err() != nil:
		run.State, run.Message = model.StateStopped, "stopped at the prompt"
	case run.State == model.StateOK:
		run.State, run.Message = model.StateFailed, "the prompt: "+err.Error()
	}
}

// step starts one job and waits for it. When the job is already
// running, step waits for that run and then starts the job again; a stop ends the run in progress.
func (r *Runner) step(ctx context.Context, step Step, user *model.User) (state, runID, message string) {
	for {
		if ctx.Err() != nil {
			return model.StateStopped, "", ""
		}
		// The run's own writes must not fail on a stop; a stop reaches the
		// run through the manager instead.
		session, err := r.manager.Start(context.WithoutCancel(ctx), runner.Request{Folder: step.Folder, Job: step.Job, User: user})
		if errors.Is(err, runner.ErrBusy) && session != nil {
			select {
			case <-session.Done():
				continue
			case <-ctx.Done():
				return model.StateStopped, "", ""
			}
		}
		if err != nil {
			return model.StateFailed, "", err.Error()
		}
		select {
		case <-session.Done():
		case <-ctx.Done():
			_ = r.manager.Stop(session.ID())
			<-session.Done()
		}
		ended := session.Snapshot()
		return ended.State, ended.ID, ""
	}
}

// nested runs the workflow a step names and waits for it. When that
// workflow is already running, nested waits for it and then starts it
// again; a stop ends the run in
// progress. The run's own state and message are the step's.
func (r *Runner) nested(ctx context.Context, wf *model.Workflow, user *model.User) (state, runID, message string) {
	for {
		if ctx.Err() != nil {
			return model.StateStopped, "", ""
		}
		run, err := r.Start(context.WithoutCancel(ctx), wf, user, "")
		if errors.Is(err, ErrBusy) {
			running, _ := r.Active(wf.ID)
			select {
			case <-r.Done(running):
				continue
			case <-ctx.Done():
				return model.StateStopped, "", ""
			}
		}
		if err != nil {
			return model.StateFailed, "", err.Error()
		}
		select {
		case <-r.Done(run.ID):
		case <-ctx.Done():
			_ = r.Stop(run.ID)
			<-r.Done(run.ID)
		}
		ended, err := r.runs.Get(context.Background(), run.ID)
		if err != nil {
			return model.StateFailed, run.ID, err.Error()
		}
		return ended.State, run.ID, ended.Message
	}
}

// saveStep writes a step's row back and tells the workflow's topic.
func (r *Runner) saveStep(ctx context.Context, row *model.WorkflowRunStep) {
	if err := r.runs.UpdateStep(ctx, row); err != nil {
		slog.Warn("workflow step not saved", "folder", row.Folder, "job", row.Job, "err", err)
	}
	r.publish(r.workflowOf(row), "step", StepEvent{
		RunID: row.WorkflowRunID, Position: row.Position, Folder: row.Folder, Job: row.Job, WorkflowRef: row.WorkflowRef,
		State: row.State, StepRunID: row.RunID, WorkflowRunRef: row.WorkflowRunRef,
	})
}

// workflowOf returns the workflow of a run's step, from the run in
// progress.
func (r *Runner) workflowOf(row *model.WorkflowRunStep) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.byRun[row.WorkflowRunID]; ok {
		return a.run.WorkflowID
	}
	return ""
}

func label(step Step) string {
	if step.Workflow != nil {
		return step.Workflow.Name
	}
	return step.Folder.ID + " / " + step.Job.ID
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
