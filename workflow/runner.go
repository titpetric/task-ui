// Package workflow runs workflows: named chains of jobs from any folders,
// started one after another as ordinary runs. A step starts when the one
// before it ended ok; the first that does not end ok ends the workflow
// run, and the runner skips the steps after it. A step names a job or
// another workflow; a workflow runs as a run of its own, and ends the
// step as it ended. The runner skips a disabled step and runs the rest.
// A step may ask Claude a prompt in place of a job, with the output of
// the steps before it; Claude's answer is the step's output for the
// prompts after it. A summary step asks Claude for task-ui's summary of
// the steps before it, and runs after a step that failed too, to say
// why. The run's answer, the reply to its input, comes from the last
// prompt or summary step that answered.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// ErrBusy reports a workflow that already has a run in progress.
var ErrBusy = errors.New("the workflow is running")

// ErrNotRunning refuses a stop of a workflow run that ended.
var ErrNotRunning = errors.New("the workflow run is not running")

// ErrHostStep refuses a step of @host:<host> without a command.
var ErrHostStep = errors.New("a step of @host:<host> needs a command as its args")

// ErrNothing refuses a run of a workflow with no enabled step.
var ErrNothing = errors.New("the workflow has no enabled step")

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
	// Asks is a prompt or a summary step, which asks Claude in Place, a
	// folder and a directory in it or _ for the root; Summary is a
	// summary step.
	Asks    bool
	Summary bool
	Place   string
	// Args follow the job's argv after --, for a job that takes args;
	// for a step of @host:<host> they are the shell command.
	Args []string
}

// Prompter puts a workflow's prompt and summary steps to Claude, as the
// user who started the run, with the run's input and the steps before
// it, in the step's place. AskStep returns when Claude replied, with the
// prompt's run and the reply, or when the prompt failed; a stop of ctx
// ends the prompt's run. note.Board is one.
type Prompter interface {
	AskStep(ctx context.Context, user *model.User, run *model.WorkflowRun, step *model.WorkflowRunStep, place, input string) (string, string, error)
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
	// Prompter puts the prompt and summary steps to Claude; nil fails them.
	Prompter Prompter
	// Hosts start a step of @host:<host> on its host. Nil fails such a step.
	Hosts *hostrun.Starter

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
		if step.IsPrompt() {
			resolved = append(resolved, Step{Asks: true, Summary: step.IsSummary(), Place: step.Place()})
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
		args := model.DecodeList(step.Args)
		if job.IsHost() && strings.TrimSpace(strings.Join(args, " ")) == "" {
			return nil, fmt.Errorf("%s: %w", step.Job, ErrHostStep)
		}
		resolved = append(resolved, Step{Folder: f, Job: job, Args: args})
	}
	return resolved, nil
}

// Start begins a run of the workflow as the user, and walks its steps in
// the background; input goes to its prompts. The caller checks the user
// may run every step.
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
	run := &model.WorkflowRun{WorkflowID: workflow.ID, Name: workflow.Name, UserID: user.ID, UserName: user.Name, Input: input}
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
	r.publish(workflow.ID, "run", RunEvent{RunID: run.ID, WorkflowID: workflow.ID, Name: run.Name, State: run.State, UserName: run.UserName, Input: input})
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
	// Input is what the run started with, on `run`; Answer is its answer, on `done`.
	Input  string `json:"input,omitempty"`
	Answer string `json:"answer,omitempty"`
}

// StepEvent is the payload of a workflow's `step` event: the step as
// its row stands.
type StepEvent struct {
	RunID       string `json:"run_id"`
	Position    int64  `json:"position"`
	Folder      string `json:"folder,omitempty"`
	Job         string `json:"job,omitempty"`
	WorkflowRef string `json:"workflow_ref,omitempty"`
	// Kind is the step's kind; Prompt says it asks Claude in Folder, a
	// prompt or a summary step.
	Kind           string `json:"kind"`
	Prompt         bool   `json:"prompt,omitempty"`
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
		if skips(run.State, step) {
			row.State = model.StepSkipped
			r.saveStep(store, row)
			continue
		}
		row.State = model.StateRunning
		r.saveStep(store, row)
		var state, message string
		switch {
		case step.Asks:
			state, row.RunID, row.Answer, message = r.promptStep(ctx, a, row, step, user)
		case step.Workflow != nil:
			state, row.WorkflowRunRef, message = r.nested(ctx, step.Workflow, user)
		default:
			// The step's row names its run once it starts, so the
			// run window follows the workflow from run to run.
			state, row.RunID, message = r.step(ctx, step, user, func(runID string) {
				row.RunID = runID
				r.saveStep(store, row)
			})
		}
		row.State = state
		r.saveStep(store, row)
		switch {
		case state == model.StateOK:
		case state == model.StateStopped && ctx.Err() != nil:
			run.State, run.Message = model.StateStopped, "stopped at "+label(step)
		case run.State == model.StateOK:
			run.State, run.Message = model.StateFailed, fmt.Sprintf("%s %s", label(step), firstOf(message, state))
		}
	}
	run.AnswerFrom(rows)
	if err := r.runs.Update(store, run); err != nil {
		slog.Warn("workflow run not saved", "workflow", run.Name, "err", err)
	}
	slog.Info("workflow finished", "workflow", run.Name, "user", run.UserName, "state", run.State)
	r.publish(run.WorkflowID, "done", RunEvent{RunID: run.ID, WorkflowID: run.WorkflowID, Name: run.Name, State: run.State, Message: run.Message, UserName: run.UserName, Answer: run.Answer})
}

// step starts one job and waits for it, and tells started the run once
// it starts. When the job is already running, step waits for that run
// and then starts the job again; a stop ends the run in progress.
func (r *Runner) step(ctx context.Context, step Step, user *model.User, started func(runID string)) (state, runID, message string) {
	if step.Job.IsHost() {
		return r.hosts(ctx, step, user, started)
	}
	for {
		if ctx.Err() != nil {
			return model.StateStopped, "", ""
		}
		// The run's own writes must not fail on a stop; a stop reaches the
		// run through the manager instead.
		session, err := r.manager.Start(context.WithoutCancel(ctx), runner.Request{Folder: step.Folder, Job: step.Job, User: user, Args: step.Args})
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
		started(session.ID())
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

// hosts starts a step of @host:<host> and waits for its runs: one on
// the host, or a batch for *. The step ends ok when every run did, else as the
// first that did not, and keeps the first run; a stop ends the runs.
func (r *Runner) hosts(ctx context.Context, step Step, user *model.User, started func(runID string)) (state, runID, message string) {
	if r.Hosts == nil {
		return model.StateFailed, "", "no hosts on this instance"
	}
	target := step.Job.HostTarget()
	req := hostrun.Request{Targets: []string{target}, Command: strings.Join(step.Args, " "), Batch: target == "*"}
	_, sessions, err := r.Hosts.Start(context.WithoutCancel(ctx), user, req)
	if err != nil {
		return model.StateFailed, "", err.Error()
	}
	started(sessions[0].ID())
	state = model.StateOK
	for _, session := range sessions {
		select {
		case <-session.Done():
		case <-ctx.Done():
			_ = r.manager.Stop(session.ID())
			<-session.Done()
		}
		ended := session.Snapshot()
		if state == model.StateOK && ended.State != model.StateOK {
			state, message = ended.State, "on "+ended.Target
		}
	}
	return state, sessions[0].ID(), message
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
		Kind: row.StepKind(), Prompt: row.IsPrompt(), State: row.State, StepRunID: row.RunID, WorkflowRunRef: row.WorkflowRunRef,
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

// skips reports whether the walk skips a step, with the run's state so
// far: a disabled one, any after a stop, and after a failure any but a
// summary.
func skips(state string, step Step) bool {
	switch {
	case step.Disabled:
		return true
	case state == model.StateOK:
		return false
	}
	return state != model.StateFailed || !step.Summary
}

func label(step Step) string {
	switch {
	case step.Summary:
		return "the summary in " + step.Place
	case step.Asks:
		return "the prompt in " + step.Place
	}
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
