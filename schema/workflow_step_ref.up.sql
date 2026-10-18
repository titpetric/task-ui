-- A step runs another workflow in place of a job: workflow_ref names it,
-- and folder and job are empty. A run's step keeps the run it made of it.
ALTER TABLE workflow_step ADD COLUMN workflow_ref CHAR(26) NOT NULL DEFAULT '';
ALTER TABLE workflow_run_step ADD COLUMN workflow_ref CHAR(26) NOT NULL DEFAULT '';
ALTER TABLE workflow_run_step ADD COLUMN workflow_run_ref CHAR(26) NOT NULL DEFAULT '';
