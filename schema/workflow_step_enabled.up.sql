-- disabled keeps a step in its workflow without running it: a run
-- records it skipped. Off, the default, the step runs.
ALTER TABLE workflow_step ADD COLUMN disabled BOOLEAN NOT NULL DEFAULT 0;
