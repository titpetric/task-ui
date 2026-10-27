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
		return errors.New("a workflow needs an enabled step")
	}
	for i, step := range steps {
		if err := checkStep(step); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// checkStep says what is wrong with one step for its kind: a job names
// its folder and job, a workflow the workflow alone, a prompt its text
// and its place, a folder or none, and a summary its place and no text.
func checkStep(step model.WorkflowStep) error {
	job := step.Folder != "" || step.Job != ""
	asks := step.Job == model.PromptJob && step.WorkflowRef == ""
	switch step.StepKind() {
	case model.StepJob:
		if step.Folder == "" || step.Job == "" || step.WorkflowRef != "" || step.Prompt != "" {
			return errors.New("a job step names a folder and a job")
		}
	case model.StepWorkflow:
		if step.WorkflowRef == "" || job || step.Prompt != "" {
			return errors.New("a workflow step names a workflow, and no job")
		}
	case model.StepPrompt:
		if strings.TrimSpace(step.Prompt) == "" || !asks {
			return errors.New("a prompt step names the folder it asks in and its prompt, and no job or workflow")
		}
	case model.StepSummary:
		if step.Prompt != "" || !asks {
			return errors.New("a summary step names the folder it asks in, and no prompt, job or workflow")
		}
	default:
		return fmt.Errorf("no step kind %q", step.Kind)
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
		row := model.WorkflowStep{ID: model.NewID(), WorkflowID: workflowID, Position: int64(i), Kind: step.StepKind(), Folder: step.Folder, Job: step.Job, WorkflowRef: step.WorkflowRef, Disabled: step.Disabled, Target: step.Target, Args: step.Args, Prompt: step.Prompt, Dir: step.Dir}
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

// Steps returns a workflow's steps in order, an empty list for a workflow without any.
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

// UsedBy returns, by name, the workflows with a step that runs the workflow.
func (s *WorkflowStorage) UsedBy(ctx context.Context, id string) ([]model.Workflow, error) {
	return client(s.db).Select[model.Workflow](ctx, "SELECT w.* FROM workflow w WHERE w.id IN (SELECT workflow_id FROM workflow_step WHERE workflow_ref = ?) ORDER BY w.name COLLATE NOCASE, w.id", id)
}

// Reaches reports whether a workflow runs the target through its steps; a workflow reaches itself.
func (s *WorkflowStorage) Reaches(ctx context.Context, from, target string) (bool, error) {
	return reaches(ctx, client(s.db), from, target, map[string]bool{})
}

// Expand replaces each workflow step with that workflow's steps at any depth, dropping a missing one.
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

// Delete removes a workflow and its steps, keeping its runs, or reports ErrInUse for one another workflow runs.
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

// Create inserts a running workflow run with its own copy of the steps, all pending.
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
				Kind: step.StepKind(), Folder: step.Folder, Job: step.Job, WorkflowRef: step.WorkflowRef, Prompt: step.Prompt, State: model.StepPending,
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

// Steps returns a workflow run's steps in order, an empty list for a run without any.
func (s *WorkflowRunStorage) Steps(ctx context.Context, id string) ([]model.WorkflowRunStep, error) {
	steps, err := client(s.db).Select[model.WorkflowRunStep](ctx, "SELECT * FROM workflow_run_step WHERE workflow_run_id = ? ORDER BY position", id)
	if steps == nil && err == nil {
		steps = []model.WorkflowRunStep{}
	}
	return steps, err
}

// StepOf returns the step of a workflow run that started a run, or
// ErrNotFound for a run no step started.
func (s *WorkflowRunStorage) StepOf(ctx context.Context, runID string) (*model.WorkflowRunStep, error) {
	if runID == "" {
		return nil, ErrNotFound
	}
	step, err := client(s.db).Get[model.WorkflowRunStep](ctx, "SELECT * FROM workflow_run_step WHERE run_id = ? ORDER BY rowid LIMIT 1", runID)
	return step, notFound(err)
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

// History returns every workflow run, oldest first, for the data retention rules.
func (s *WorkflowRunStorage) History(ctx context.Context) ([]model.WorkflowRun, error) {
	return client(s.db).Select[model.WorkflowRun](ctx, "SELECT * FROM workflow_run ORDER BY started_at ASC, id ASC")
}

// Delete removes a workflow run with its steps; the workflow stays.
func (s *WorkflowRunStorage) Delete(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM workflow_run_step WHERE workflow_run_id = ?", id); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM workflow_run WHERE id = ?", id)
	})
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

// Interrupt fails the runs a stopped instance left running and their running step, skipping the steps after it.
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

// Answered returns at most limit finished runs of a workflow with an answer, newest first, except the run it names.
func (s *WorkflowRunStorage) Answered(ctx context.Context, workflowID, except string, limit int) ([]model.WorkflowRun, error) {
	return client(s.db).Select[model.WorkflowRun](ctx,
		"SELECT * FROM workflow_run WHERE workflow_id = ? AND id != ? AND answer != '' AND state != ? ORDER BY started_at DESC, rowid DESC LIMIT ?",
		workflowID, except, model.StateRunning, limit)
}

// Number returns the place of a run among its workflow's runs, the
// first 1, in the order they started.
func (s *WorkflowRunStorage) Number(ctx context.Context, run *model.WorkflowRun) (int, error) {
	n, err := client(s.db).Get[int64](ctx, "SELECT COUNT(*) FROM workflow_run WHERE workflow_id = ? AND rowid <= (SELECT rowid FROM workflow_run WHERE id = ?)", run.WorkflowID, run.ID)
	if err != nil {
		return 0, err
	}
	return int(*n), nil
}
