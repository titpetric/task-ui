package storage

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
)

// ErrInUse reports deleting a workflow another one runs.
var ErrInUse = errors.New("the workflow is in use")

// WorkflowStorage keeps the workflows and their steps.
type WorkflowStorage struct {
	db DB
}

// NewWorkflowStorage returns the storage over a pool.
func NewWorkflowStorage(db DB) *WorkflowStorage {
	return &WorkflowStorage{db: db}
}

// checkChain says what is wrong with a workflow's name and steps.
func checkChain(workflow *model.Workflow, steps []model.WorkflowStep) error {
	if strings.TrimSpace(workflow.Name) == "" {
		return errors.New("a workflow needs a name")
	}
	if !workflow.Runs(steps) {
		return errors.New("a workflow needs a step to run or a prompt")
	}
	for _, step := range steps {
		job := step.Folder != "" || step.Job != ""
		switch {
		case step.IsPrompt() && (step.IsWorkflow() || step.Folder == "" || step.Job != model.PromptJob):
			return errors.New("a prompt step names the folder it asks in, and no job or workflow")
		case step.IsWorkflow() && job:
			return errors.New("a step names a job or a workflow, not both")
		case !step.IsWorkflow() && (step.Folder == "" || step.Job == ""):
			return errors.New("every step names a folder and a job, or a workflow")
		}
	}
	return nil
}

// checkRefs says what is wrong with the workflows the steps run: one
// that is not there, or one that runs the workflow itself, a loop.
func checkRefs(ctx context.Context, c conn, workflowID string, steps []model.WorkflowStep) error {
	for _, step := range steps {
		if !step.IsWorkflow() {
			continue
		}
		if step.WorkflowRef == workflowID {
			return errors.New("a workflow may not run itself: a loop")
		}
		ref, err := c.Get[model.Workflow](ctx, "SELECT * FROM workflow WHERE id = ?", step.WorkflowRef)
		if errors.Is(notFound(err), ErrNotFound) {
			return fmt.Errorf("no workflow %s", step.WorkflowRef)
		}
		if err != nil {
			return err
		}
		loops, err := reaches(ctx, c, ref.ID, workflowID, map[string]bool{})
		if err != nil {
			return err
		}
		if loops {
			return fmt.Errorf("%s runs this workflow: a loop", ref.Name)
		}
	}
	return nil
}

// reaches reports whether a workflow runs the target, through any
// workflows its steps run.
func reaches(ctx context.Context, c conn, from, target string, seen map[string]bool) (bool, error) {
	if from == target {
		return true, nil
	}
	if seen[from] {
		return false, nil
	}
	seen[from] = true
	refs, err := c.Select[model.WorkflowStep](ctx, "SELECT * FROM workflow_step WHERE workflow_id = ? AND workflow_ref != '' ORDER BY position", from)
	if err != nil {
		return false, err
	}
	for _, ref := range refs {
		found, err := reaches(ctx, c, ref.WorkflowRef, target, seen)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

// inTx runs fn in a transaction on one client, committing when it
// returns nil.
func inTx(ctx context.Context, db DB, fn func(c conn) error) error {
	c := client(db)
	if err := c.Begin(ctx); err != nil {
		return err
	}
	if err := fn(c); err != nil {
		return errors.Join(err, c.Rollback())
	}
	return c.Commit()
}

// insertSteps writes a workflow's steps in order, positions from zero.
func insertSteps(ctx context.Context, c conn, workflowID string, steps []model.WorkflowStep) error {
	for i, step := range steps {
		row := model.WorkflowStep{ID: model.NewID(), WorkflowID: workflowID, Position: int64(i), Folder: step.Folder, Job: step.Job, WorkflowRef: step.WorkflowRef, Disabled: step.Disabled, Target: step.Target, Args: step.Args, Prompt: step.Prompt, Dir: step.Dir}
		if row.Args == "" {
			row.Args = "[]"
		}
		if err := c.Insert(ctx, model.WorkflowStepTable, &row); err != nil {
			return err
		}
	}
	return nil
}

// Create inserts a workflow and its steps, in the order given.
func (s *WorkflowStorage) Create(ctx context.Context, workflow *model.Workflow, steps []model.WorkflowStep) error {
	if err := checkChain(workflow, steps); err != nil {
		return err
	}
	if workflow.ID == "" {
		workflow.ID = model.NewID()
	}
	workflow.CreatedAt = stamp()
	workflow.UpdatedAt = workflow.CreatedAt
	return inTx(ctx, s.db, func(c conn) error {
		if err := checkRefs(ctx, c, workflow.ID, steps); err != nil {
			return err
		}
		if err := c.Insert(ctx, model.WorkflowTable, workflow); err != nil {
			return err
		}
		return insertSteps(ctx, c, workflow.ID, steps)
	})
}

// Get returns a workflow by ID.
func (s *WorkflowStorage) Get(ctx context.Context, id string) (*model.Workflow, error) {
	workflow, err := client(s.db).Get[model.Workflow](ctx, "SELECT * FROM workflow WHERE id = ?", id)
	return workflow, notFound(err)
}

// Steps returns a workflow's steps in order, an empty list for a
// workflow of a prompt alone.
func (s *WorkflowStorage) Steps(ctx context.Context, id string) ([]model.WorkflowStep, error) {
	steps, err := client(s.db).Select[model.WorkflowStep](ctx, "SELECT * FROM workflow_step WHERE workflow_id = ? ORDER BY position", id)
	if steps == nil && err == nil {
		steps = []model.WorkflowStep{}
	}
	return steps, err
}

// List returns every workflow, by name.
func (s *WorkflowStorage) List(ctx context.Context) ([]model.Workflow, error) {
	return client(s.db).Select[model.Workflow](ctx, "SELECT * FROM workflow ORDER BY name COLLATE NOCASE, id")
}

// UsedBy returns the workflows with a step that runs the workflow, by
// name.
func (s *WorkflowStorage) UsedBy(ctx context.Context, id string) ([]model.Workflow, error) {
	return client(s.db).Select[model.Workflow](ctx, "SELECT w.* FROM workflow w WHERE w.id IN (SELECT workflow_id FROM workflow_step WHERE workflow_ref = ?) ORDER BY w.name COLLATE NOCASE, w.id", id)
}

// Reaches reports whether a workflow runs the target, through any
// workflows its steps run; a workflow reaches itself.
func (s *WorkflowStorage) Reaches(ctx context.Context, from, target string) (bool, error) {
	return reaches(ctx, client(s.db), from, target, map[string]bool{})
}

// Expand replaces each step that runs a workflow with that workflow's
// steps, through any depth, so a check sees every job the chain runs. Expand
// drops a step whose workflow does not exist.
func (s *WorkflowStorage) Expand(ctx context.Context, steps []model.WorkflowStep) ([]model.WorkflowStep, error) {
	return s.expand(ctx, steps, map[string]bool{})
}

func (s *WorkflowStorage) expand(ctx context.Context, steps []model.WorkflowStep, seen map[string]bool) ([]model.WorkflowStep, error) {
	flat := make([]model.WorkflowStep, 0, len(steps))
	for _, step := range steps {
		if !step.IsWorkflow() {
			flat = append(flat, step)
			continue
		}
		if seen[step.WorkflowRef] {
			continue
		}
		seen[step.WorkflowRef] = true
		inner, err := s.Steps(ctx, step.WorkflowRef)
		if err != nil {
			return nil, err
		}
		inner, err = s.expand(ctx, inner, seen)
		if err != nil {
			return nil, err
		}
		flat = append(flat, inner...)
	}
	return flat, nil
}

// Update writes a workflow's name back and replaces its steps.
func (s *WorkflowStorage) Update(ctx context.Context, workflow *model.Workflow, steps []model.WorkflowStep) error {
	if err := checkChain(workflow, steps); err != nil {
		return err
	}
	workflow.UpdatedAt = stamp()
	return inTx(ctx, s.db, func(c conn) error {
		if err := checkRefs(ctx, c, workflow.ID, steps); err != nil {
			return err
		}
		if err := c.Update(ctx, model.WorkflowTable, workflow, "id"); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM workflow_step WHERE workflow_id = ?", workflow.ID); err != nil {
			return err
		}
		return insertSteps(ctx, c, workflow.ID, steps)
	})
}

// Delete removes a workflow and its steps. Its runs stay, as history. A
// workflow another one runs stays, with ErrInUse.
func (s *WorkflowStorage) Delete(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c conn) error {
		users, err := c.Select[model.Workflow](ctx, "SELECT w.* FROM workflow w WHERE w.id IN (SELECT workflow_id FROM workflow_step WHERE workflow_ref = ?) ORDER BY w.name COLLATE NOCASE", id)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			names := make([]string, 0, len(users))
			for _, u := range users {
				names = append(names, u.Name)
			}
			return fmt.Errorf("%w: %s runs it", ErrInUse, strings.Join(names, ", "))
		}
		if err := c.Exec(ctx, "DELETE FROM workflow_step WHERE workflow_id = ?", id); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM workflow WHERE id = ?", id)
	})
}

// WorkflowRunStorage keeps the runs of workflows and their steps.
type WorkflowRunStorage struct {
	db DB
}

// NewWorkflowRunStorage returns the storage over a pool.
func NewWorkflowRunStorage(db DB) *WorkflowRunStorage {
	return &WorkflowRunStorage{db: db}
}

// Create inserts a running workflow run and its steps, all pending. Create
// copies the steps, so a later change to the workflow leaves the run's
// history intact.
func (s *WorkflowRunStorage) Create(ctx context.Context, run *model.WorkflowRun, steps []model.WorkflowStep) error {
	if run.ID == "" {
		run.ID = model.NewID()
	}
	run.State = model.StateRunning
	run.StartedAt = stamp()
	run.CreatedAt, run.UpdatedAt = run.StartedAt, run.StartedAt
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Insert(ctx, model.WorkflowRunTable, run); err != nil {
			return err
		}
		for i, step := range steps {
			row := model.WorkflowRunStep{
				ID: model.NewID(), WorkflowRunID: run.ID, Position: int64(i),
				Folder: step.Folder, Job: step.Job, WorkflowRef: step.WorkflowRef, Prompt: step.Prompt, State: model.StepPending,
				CreatedAt: run.StartedAt, UpdatedAt: run.StartedAt,
			}
			if err := c.Insert(ctx, model.WorkflowRunStepTable, &row); err != nil {
				return err
			}
		}
		return nil
	})
}

// Get returns a workflow run by ID.
func (s *WorkflowRunStorage) Get(ctx context.Context, id string) (*model.WorkflowRun, error) {
	run, err := client(s.db).Get[model.WorkflowRun](ctx, "SELECT * FROM workflow_run WHERE id = ?", id)
	return run, notFound(err)
}

// Steps returns a workflow run's steps in order, an empty list for a
// run of a prompt alone.
func (s *WorkflowRunStorage) Steps(ctx context.Context, id string) ([]model.WorkflowRunStep, error) {
	steps, err := client(s.db).Select[model.WorkflowRunStep](ctx, "SELECT * FROM workflow_run_step WHERE workflow_run_id = ? ORDER BY position", id)
	if steps == nil && err == nil {
		steps = []model.WorkflowRunStep{}
	}
	return steps, err
}

// WorkflowRunFilter pages a workflow's runs: Limit of them, 20 when
// zero, older than the run Before names or newer than the one After
// names.
type WorkflowRunFilter struct {
	Limit  int
	Before string
	After  string
}

// ListByWorkflow returns a workflow's runs, newest first.
func (s *WorkflowRunStorage) ListByWorkflow(ctx context.Context, workflowID string, filter WorkflowRunFilter) ([]model.WorkflowRun, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	query := "SELECT * FROM workflow_run WHERE workflow_id = ?"
	args := []any{workflowID}
	page, pageArgs, order := cursor("workflow_run", filter.Before, filter.After)
	if page != "" {
		query += " AND " + page
		args = append(args, pageArgs...)
	}
	args = append(args, filter.Limit)
	runs, err := client(s.db).Select[model.WorkflowRun](ctx, query+" ORDER BY "+order+" LIMIT ?", args...)
	if filter.Before == "" && filter.After != "" {
		slices.Reverse(runs)
	}
	return runs, err
}

// List returns the runs of every workflow, newest first.
func (s *WorkflowRunStorage) List(ctx context.Context, limit int) ([]model.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	return client(s.db).Select[model.WorkflowRun](ctx, "SELECT * FROM workflow_run ORDER BY started_at DESC, id DESC LIMIT ?", limit)
}

// Update writes a workflow run back, by ID; a finished state stamps it.
func (s *WorkflowRunStorage) Update(ctx context.Context, run *model.WorkflowRun) error {
	run.UpdatedAt = stamp()
	if model.Finished(run.State) && run.FinishedAt == nil {
		run.FinishedAt = run.UpdatedAt
	}
	return client(s.db).Update(ctx, model.WorkflowRunTable, run, "id")
}

// UpdateStep writes a step of a workflow run back, by ID.
func (s *WorkflowRunStorage) UpdateStep(ctx context.Context, step *model.WorkflowRunStep) error {
	step.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.WorkflowRunStepTable, step, "id")
}

// Interrupt ends the workflow runs left running, when the instance that
// ran them has stopped: they fail with the message, their running step
// fails and Interrupt skips the steps after it.
func (s *WorkflowRunStorage) Interrupt(ctx context.Context, message string) error {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "UPDATE workflow_run_step SET state = ?, updated_at = ? WHERE state = ? AND workflow_run_id IN (SELECT id FROM workflow_run WHERE state = ?)", model.StateFailed, now, model.StateRunning, model.StateRunning); err != nil {
			return err
		}
		if err := c.Exec(ctx, "UPDATE workflow_run_step SET state = ?, updated_at = ? WHERE state = ? AND workflow_run_id IN (SELECT id FROM workflow_run WHERE state = ?)", model.StepSkipped, now, model.StepPending, model.StateRunning); err != nil {
			return err
		}
		return c.Exec(ctx, "UPDATE workflow_run SET state = ?, message = ?, finished_at = ?, updated_at = ? WHERE state = ?", model.StateFailed, message, now, now, model.StateRunning)
	})
}
