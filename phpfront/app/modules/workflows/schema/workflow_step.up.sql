-- workflow_step: One job of a workflow, at its position in the chain
CREATE TABLE IF NOT EXISTS workflow_step (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workflow_id CHAR(26) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    folder TEXT NOT NULL,
    job TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_workflow_step_workflow ON workflow_step(workflow_id, position);
