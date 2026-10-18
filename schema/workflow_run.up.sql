-- workflow_run: One run of a workflow, its steps started one after another
CREATE TABLE IF NOT EXISTS workflow_run (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workflow_id CHAR(26) NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'running' CHECK (state IN ('running', 'ok', 'failed', 'stopped')),
    message TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_workflow_run_workflow ON workflow_run(workflow_id, started_at);
CREATE INDEX IF NOT EXISTS idx_workflow_run_state ON workflow_run(state);
