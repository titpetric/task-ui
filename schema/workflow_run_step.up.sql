-- workflow_run_step: One step of a workflow run: the job it names and the run that did it
CREATE TABLE IF NOT EXISTS workflow_run_step (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workflow_run_id CHAR(26) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    run_id CHAR(26) NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'running', 'ok', 'failed', 'stopped', 'killed', 'skipped')),
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_workflow_run_step_run ON workflow_run_step(workflow_run_id, position);
