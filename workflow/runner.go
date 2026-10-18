// Package workflow runs workflows: named chains of jobs from any folders,
// started one after another as ordinary runs. A step starts when the one
// before it ended ok; the first that does not end ok ends the workflow
// run, and the steps after it are skipped.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// ErrBusy is returned when the workflow already has a run in progress.
var ErrBusy = errors.New("the workflow is running")

// ErrNotRunning is returned when stopping a workflow run that ended.
var ErrNotRunning = errors.New("the workflow run is not running")

// Catalogue finds the folder a step names. folder.Discovery is one.
type Catalogue interface {
	Find(id string) *folder.Folder
}

// Step is a workflow step resolved to its folder and job.
type Step struct {
	Folder *folder.Folder
	Job    *folder.Job
}

// Runner starts workflow runs and walks their steps.
type Runner struct {
	workflows *storage.WorkflowStorage
	runs      *storage.WorkflowRunStorage
	catalogue Catalogue
	manager   *runner.Manager

	mu     sync.Mutex
	byFlow map[string]*active // workflow ID -> its run in progress
	byRun  map[string]*active // workflow run ID -> the same
}

// active is a workflow run in progress.
type active struct {
	run    *model.WorkflowRun
	cancel context.CancelFunc
	done   chan struct{}
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

// Resolve finds the folder and job of every step, or says which one is
// gone.
func (r *Runner) Resolve(steps []model.WorkflowStep) ([]Step, error) {
	resolved := make([]Step, 0, len(steps))
	for _, step := range steps {
		f := r.catalogue.Find(step.Folder)
		var job *folder.Job
		if f != nil {
			job = f.Job(step.Job)
		}
		if job == nil {
			return nil, fmt.Errorf("no job %s / %s", step.Folder, step.Job)
		}
		resolved = append(resolved, Step{Folder: f, Job: job})
	}
	return resolved, nil
}

// Start begins a run of the workflow as the user, and walks its steps in
// the background. The caller checks the user may run every step.
func (r *Runner) Start(ctx context.Context, workflow *model.Workflow, user *model.User) (*model.WorkflowRun, error) {
	steps, err := r.workflows.Steps(ctx, workflow.ID)
	if err != nil {
		return nil, err
	}
	resolved, err := r.Resolve(steps)
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
	a := &active{run: run, cancel: cancel, done: make(chan struct{})}
	r.byFlow[workflow.ID], r.byRun[run.ID] = a, a
	snapshot := *run
	go r.walk(walkCtx, a, resolved, rows, user)
	return &snapshot, nil
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

// Done is closed when the workflow run ended; for a run not in progress
// it is closed already.
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

// Stop ends a workflow run: the step running is stopped, the rest are
// skipped.
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
	// The records outlive a stopped walk, so they are written without its
	// context.
	store := context.Background()
	run := a.run
	run.State = model.StateOK

	for i, step := range steps {
		row := &rows[i]
		if run.State != model.StateOK {
			row.State = model.StepSkipped
			r.saveStep(store, row)
			continue
		}
		state, runID, message := r.step(ctx, step, user)
		row.State, row.RunID = state, runID
		r.saveStep(store, row)
		switch {
		case state == model.StateOK:
		case state == model.StateStopped && ctx.Err() != nil:
			run.State, run.Message = model.StateStopped, "stopped at "+label(step)
		default:
			run.State, run.Message = model.StateFailed, fmt.Sprintf("%s %s", label(step), firstOf(message, state))
		}
	}
	if err := r.runs.Update(store, run); err != nil {
		slog.Warn("workflow run not saved", "workflow", run.Name, "err", err)
	}
	slog.Info("workflow finished", "workflow", run.Name, "user", run.UserName, "state", run.State)
}

// step starts one job and waits for it. A job already running is waited
// for and then started again; a stop ends the run in progress.
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

func (r *Runner) saveStep(ctx context.Context, row *model.WorkflowRunStep) {
	if err := r.runs.UpdateStep(ctx, row); err != nil {
		slog.Warn("workflow step not saved", "folder", row.Folder, "job", row.Job, "err", err)
	}
}

func label(step Step) string {
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
