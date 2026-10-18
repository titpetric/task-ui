package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/titpetric/pdo"

	"github.com/titpetric/task-ui/model"
)

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
	if len(steps) == 0 {
		return errors.New("a workflow needs at least one step")
	}
	for _, step := range steps {
		if step.Folder == "" || step.Job == "" {
			return errors.New("every step names a folder and a job")
		}
	}
	return nil
}

// inTx runs fn in a transaction on one client, committing when it
// returns nil.
func inTx(ctx context.Context, db DB, fn func(c *pdo.PDO) error) error {
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
func insertSteps(ctx context.Context, c *pdo.PDO, workflowID string, steps []model.WorkflowStep) error {
	for i, step := range steps {
		row := model.WorkflowStep{ID: model.NewID(), WorkflowID: workflowID, Position: int64(i), Folder: step.Folder, Job: step.Job}
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
	return inTx(ctx, s.db, func(c *pdo.PDO) error {
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

// Steps returns a workflow's steps in order.
func (s *WorkflowStorage) Steps(ctx context.Context, id string) ([]model.WorkflowStep, error) {
	return client(s.db).Select[model.WorkflowStep](ctx, "SELECT * FROM workflow_step WHERE workflow_id = ? ORDER BY position", id)
}

// List returns every workflow, by name.
func (s *WorkflowStorage) List(ctx context.Context) ([]model.Workflow, error) {
	return client(s.db).Select[model.Workflow](ctx, "SELECT * FROM workflow ORDER BY name COLLATE NOCASE, id")
}

// Update writes a workflow's name back and replaces its steps.
func (s *WorkflowStorage) Update(ctx context.Context, workflow *model.Workflow, steps []model.WorkflowStep) error {
	if err := checkChain(workflow, steps); err != nil {
		return err
	}
	workflow.UpdatedAt = stamp()
	return inTx(ctx, s.db, func(c *pdo.PDO) error {
		if err := c.Update(ctx, model.WorkflowTable, workflow, "id"); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM workflow_step WHERE workflow_id = ?", workflow.ID); err != nil {
			return err
		}
		return insertSteps(ctx, c, workflow.ID, steps)
	})
}

// Delete removes a workflow and its steps. Its runs stay, as history.
func (s *WorkflowStorage) Delete(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c *pdo.PDO) error {
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

// Create inserts a running workflow run and its steps, all pending. The
// steps are copied, so a later change to the workflow leaves the run's
// history as it was.
func (s *WorkflowRunStorage) Create(ctx context.Context, run *model.WorkflowRun, steps []model.WorkflowStep) error {
	if run.ID == "" {
		run.ID = model.NewID()
	}
	run.State = model.StateRunning
	run.StartedAt = stamp()
	run.CreatedAt, run.UpdatedAt = run.StartedAt, run.StartedAt
	return inTx(ctx, s.db, func(c *pdo.PDO) error {
		if err := c.Insert(ctx, model.WorkflowRunTable, run); err != nil {
			return err
		}
		for i, step := range steps {
			row := model.WorkflowRunStep{
				ID: model.NewID(), WorkflowRunID: run.ID, Position: int64(i),
				Folder: step.Folder, Job: step.Job, State: model.StepPending,
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

// Steps returns a workflow run's steps in order.
func (s *WorkflowRunStorage) Steps(ctx context.Context, id string) ([]model.WorkflowRunStep, error) {
	return client(s.db).Select[model.WorkflowRunStep](ctx, "SELECT * FROM workflow_run_step WHERE workflow_run_id = ? ORDER BY position", id)
}

// ListByWorkflow returns a workflow's runs, newest first.
func (s *WorkflowRunStorage) ListByWorkflow(ctx context.Context, workflowID string, limit int) ([]model.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	return client(s.db).Select[model.WorkflowRun](ctx, "SELECT * FROM workflow_run WHERE workflow_id = ? ORDER BY started_at DESC, id DESC LIMIT ?", workflowID, limit)
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
// ran them is gone: they fail with the message, their running step fails
// and the steps after it are skipped.
func (s *WorkflowRunStorage) Interrupt(ctx context.Context, message string) error {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return inTx(ctx, s.db, func(c *pdo.PDO) error {
		if err := c.Exec(ctx, "UPDATE workflow_run_step SET state = ?, updated_at = ? WHERE state = ? AND workflow_run_id IN (SELECT id FROM workflow_run WHERE state = ?)", model.StateFailed, now, model.StateRunning, model.StateRunning); err != nil {
			return err
		}
		if err := c.Exec(ctx, "UPDATE workflow_run_step SET state = ?, updated_at = ? WHERE state = ? AND workflow_run_id IN (SELECT id FROM workflow_run WHERE state = ?)", model.StepSkipped, now, model.StepPending, model.StateRunning); err != nil {
			return err
		}
		return c.Exec(ctx, "UPDATE workflow_run SET state = ?, message = ?, finished_at = ?, updated_at = ? WHERE state = ?", model.StateFailed, message, now, now, model.StateRunning)
	})
}
